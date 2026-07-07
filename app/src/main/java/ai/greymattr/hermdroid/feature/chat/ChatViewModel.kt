package ai.greymattr.hermdroid.feature.chat

import ai.greymattr.hermdroid.data.local.SecureSettings
import android.app.Application
import android.net.Uri
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import core.HermeyClient
import core.sse.EventListener
import core.stream.StreamManager
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.channels.awaitClose
import kotlinx.coroutines.channels.trySendBlocking
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.callbackFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.util.UUID

sealed class ChatUiState {
    object Loading : ChatUiState()
    data class Ready(val sessionTitle: String = "") : ChatUiState()
    data class Error(val message: String) : ChatUiState()
}

class ChatViewModel(application: Application) : AndroidViewModel(application) {
    private val _uiState = MutableStateFlow<ChatUiState>(ChatUiState.Ready())
    val uiState: StateFlow<ChatUiState> = _uiState.asStateFlow()

    private val _messages = MutableStateFlow<List<UiMessage>>(emptyList())
    val messages: StateFlow<List<UiMessage>> = _messages.asStateFlow()

    private val _inputText = MutableStateFlow("")
    val inputText: StateFlow<String> = _inputText.asStateFlow()

    private val _isStreaming = MutableStateFlow(false)
    val isStreaming: StateFlow<Boolean> = _isStreaming.asStateFlow()

    private val _selectedModel = MutableStateFlow(SecureSettings.defaultModel)
    val selectedModel: StateFlow<String> = _selectedModel.asStateFlow()

    private val _reasoningEnabled = MutableStateFlow(false)
    val reasoningEnabled: StateFlow<Boolean> = _reasoningEnabled.asStateFlow()

    private val _attachments = MutableStateFlow<List<Attachment>>(emptyList())
    val attachments: StateFlow<List<Attachment>> = _attachments.asStateFlow()

    private var client: HermeyClient? = null
    private var streamManager: StreamManager? = null
    private var currentStreamId: String? = null
    private var sessionId: String? = null

    fun initSession(sessionId: String, title: String = "") {
        this.sessionId = sessionId
        _uiState.value = ChatUiState.Ready(title)
        loadHistory(sessionId)
    }

    private fun ensureClient(): HermeyClient? {
        if (client != null) return client
        val url = SecureSettings.serverUrl.takeIf { it.isNotBlank() } ?: return null
        return try {
            HermeyClient(url).also { client = it }
        } catch (e: Exception) {
            _uiState.value = ChatUiState.Error("Client init failed: ${e.message}")
            null
        }
    }

    private fun loadHistory(sessionId: String) {
        viewModelScope.launch(Dispatchers.IO) {
            val c = ensureClient() ?: return@launch
            try {
                val resp = c.getChatHistory(sessionId, 100)
                resp?.messages?.let { history ->
                    val loaded = history.map { msg ->
                        UiMessage(
                            id = msg.id ?: UUID.randomUUID().toString(),
                            role = when (msg.role) {
                                "assistant" -> MessageRole.Assistant
                                "system" -> MessageRole.System
                                else -> MessageRole.User
                            },
                            content = StringBuilder(msg.content ?: "")
                        )
                    }
                    _messages.value = loaded
                }
            } catch (e: Exception) {
                // Offline or error: keep empty history.
            }
        }
    }

    fun onInputChange(text: String) {
        _inputText.value = text
    }

    fun onModelSelected(model: String) {
        _selectedModel.value = model
    }

    fun toggleReasoning() {
        _reasoningEnabled.update { !it }
    }

    fun addAttachment(attachment: Attachment) {
        _attachments.update { it + attachment }
    }

    fun removeAttachment(uri: String) {
        _attachments.update { list -> list.filter { it.uri != uri } }
    }

    fun send() {
        val text = _inputText.value.trim()
        if (text.isBlank() && _attachments.value.isEmpty()) return
        val sid = sessionId ?: run {
            _uiState.value = ChatUiState.Error("No session open")
            return
        }
        val model = _selectedModel.value

        val userMessage = UiMessage(
            id = UUID.randomUUID().toString(),
            role = MessageRole.User,
            content = StringBuilder(text),
            attachments = _attachments.value.toList()
        )
        _messages.update { it + userMessage }
        _inputText.value = ""
        _attachments.value = emptyList()

        val assistantMessage = UiMessage(
            id = UUID.randomUUID().toString(),
            role = MessageRole.Assistant,
            isStreaming = true,
            reasoning = if (_reasoningEnabled.value) UiReasoning() else null
        )
        _messages.update { it + assistantMessage }
        _isStreaming.value = true

        viewModelScope.launch(Dispatchers.IO) {
            val c = ensureClient() ?: run {
                markStreamError("Server not configured")
                return@launch
            }
            try {
                val sm = c.newStreamManager().also { streamManager = it }
                val req = core.endpoints.StartChatRequest().apply {
                    setSessionID(sid)
                    setMessage(text)
                    setWorkspace("")
                    setModel(model)
                }
                val start = sm.start(req, streamingListener { event ->
                    handleEvent(event, assistantMessage.id)
                })
                currentStreamId = start?.streamID
                if (start == null) {
                    markStreamError("Failed to start stream")
                }
            } catch (e: Exception) {
                markStreamError(e.message ?: "Stream failed")
            }
        }
    }

    fun stop() {
        val sm = streamManager ?: return
        viewModelScope.launch(Dispatchers.IO) {
            try {
                sm.cancel()
            } catch (_: Exception) {
            } finally {
                _isStreaming.value = false
                finalizeAssistantMessage()
            }
        }
    }

    fun steer(text: String) {
        val sm = streamManager ?: return
        viewModelScope.launch(Dispatchers.IO) {
            try {
                sm.steer(text)
            } catch (e: Exception) {
                _messages.update { list ->
                    list + UiMessage(
                        id = UUID.randomUUID().toString(),
                        role = MessageRole.System,
                        content = StringBuilder("Steer failed: ${e.message}")
                    )
                }
            }
        }
    }

    fun retry() {
        // Resend last user message if present.
        val lastUser = _messages.value.findLast { it.role == MessageRole.User } ?: return
        _inputText.value = lastUser.displayContent
        // Remove the failed assistant turn if it was the last message.
        _messages.update { list ->
            val last = list.lastOrNull()
            if (last?.role == MessageRole.Assistant && last.isStreaming) list.dropLast(1) else list
        }
        send()
    }

    fun toggleReasoningExpanded(messageId: String) {
        _messages.update { list ->
            list.map { msg ->
                if (msg.id == messageId && msg.reasoning != null) {
                    msg.copy(reasoning = msg.reasoning.copy(expanded = !msg.reasoning.expanded))
                } else msg
            }
        }
    }

    fun toggleToolExpanded(messageId: String, toolId: String) {
        _messages.update { list ->
            list.map { msg ->
                if (msg.id == messageId) {
                    msg.copy(toolCalls = msg.toolCalls.map { tool ->
                        if (tool.id == toolId) tool.copy(expanded = !tool.expanded) else tool
                    }.toMutableList())
                } else msg
            }
        }
    }

    private fun handleEvent(event: ChatEvent, assistantId: String) {
        _messages.update { list ->
            list.map { msg ->
                if (msg.id != assistantId) return@map msg
                when (event) {
                    is ChatEvent.Token -> msg.apply { content.append(event.text) }
                    is ChatEvent.Reasoning -> msg.apply {
                        if (reasoning == null) return@apply
                        reasoning.text.append(event.text)
                    }
                    is ChatEvent.ToolCall -> msg.apply { toolCalls.add(parseToolCall(event.json)) }
                    is ChatEvent.ToolResult -> msg.apply { attachToolResult(event.json) }
                    is ChatEvent.Error -> msg.copy(streamError = event.message, isStreaming = false)
                    is ChatEvent.StreamEnd, ChatEvent.Cancel -> msg.copy(isStreaming = false)
                    else -> msg
                }
            }
        }
        if (event is ChatEvent.StreamEnd || event is ChatEvent.Error || event is ChatEvent.Cancel) {
            _isStreaming.value = false
            finalizeAssistantMessage()
        }
    }

    private fun finalizeAssistantMessage() {
        _messages.update { list ->
            list.map { msg ->
                if (msg.role == MessageRole.Assistant && msg.isStreaming) {
                    msg.copy(isStreaming = false)
                } else msg
            }
        }
    }

    private fun markStreamError(message: String) {
        _isStreaming.value = false
        _messages.update { list ->
            val last = list.lastOrNull()
            if (last?.role == MessageRole.Assistant) {
                list.dropLast(1) + last.copy(streamError = message, isStreaming = false)
            } else list + UiMessage(
                id = UUID.randomUUID().toString(),
                role = MessageRole.System,
                content = StringBuilder(message)
            )
        }
    }

    private fun parseToolCall(json: String): UiToolCall {
        return try {
            val obj = JSONObject(json)
            UiToolCall(
                id = obj.optString("id", UUID.randomUUID().toString()),
                name = obj.optString("name", "tool"),
                input = obj.optString("input", json),
                status = ToolStatus.Running
            )
        } catch (_: Exception) {
            UiToolCall(
                id = UUID.randomUUID().toString(),
                name = "tool",
                input = json,
                status = ToolStatus.Running
            )
        }
    }

    private fun UiMessage.attachToolResult(json: String) {
        try {
            val obj = JSONObject(json)
            val id = obj.optString("id")
            val result = obj.optString("result")
            val status = obj.optString("status")
            val tool = toolCalls.find { it.id == id } ?: toolCalls.lastOrNull()
            tool?.let {
                val index = toolCalls.indexOf(it)
                if (index >= 0) {
                    toolCalls[index] = it.copy(
                        result = result,
                        status = when (status.lowercase()) {
                            "error" -> ToolStatus.Error
                            "success" -> ToolStatus.Success
                            else -> ToolStatus.Success
                        }
                    )
                }
            }
        } catch (_: Exception) {
        }
    }

    private fun streamingListener(onEvent: (ChatEvent) -> Unit): EventListener {
        return object : EventListener {
            override fun onToken(text: String?) {
                text?.let { onEvent(ChatEvent.Token(it)) }
            }
            override fun onToolCall(callJSON: String?) {
                callJSON?.let { onEvent(ChatEvent.ToolCall(it)) }
            }
            override fun onToolResult(resultJSON: String?) {
                resultJSON?.let { onEvent(ChatEvent.ToolResult(it)) }
            }
            override fun onReasoning(text: String?) {
                text?.let { onEvent(ChatEvent.Reasoning(it)) }
            }
            override fun onStreamEnd() {
                onEvent(ChatEvent.StreamEnd)
            }
            override fun onError(msg: String?) {
                onEvent(ChatEvent.Error(msg ?: "Unknown stream error"))
            }
            override fun onCancel() {
                onEvent(ChatEvent.Cancel)
            }
        }
    }
}
