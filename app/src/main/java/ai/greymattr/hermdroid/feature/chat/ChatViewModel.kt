package ai.greymattr.hermdroid.feature.chat

import ai.greymattr.hermdroid.data.local.SecureSettings
import android.app.Application
import android.net.Uri
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import core.EventListenerProxy
import core.HermeyClient
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import org.json.JSONArray
import org.json.JSONObject
import java.util.UUID
import java.util.concurrent.atomic.AtomicBoolean

sealed class ChatUiState {
    object Loading : ChatUiState()
    data class Ready(val sessionTitle: String = "") : ChatUiState()
    data class Error(val message: String) : ChatUiState()
}

class ChatViewModel(application: Application) : AndroidViewModel(application) {
    private val _uiState = MutableStateFlow<ChatUiState>(ChatUiState.Ready())
    val uiState: StateFlow<ChatUiState> = _uiState

    private val _messages = MutableStateFlow<List<UiMessage>>(emptyList())
    val messages: StateFlow<List<UiMessage>> = _messages

    private val _inputText = MutableStateFlow("")
    val inputText: StateFlow<String> = _inputText

    private val _isStreaming = MutableStateFlow(false)
    val isStreaming: StateFlow<Boolean> = _isStreaming

    private val _selectedModel = MutableStateFlow(SecureSettings.defaultModel)
    val selectedModel: StateFlow<String> = _selectedModel

    private val _reasoningEnabled = MutableStateFlow(false)
    val reasoningEnabled: StateFlow<Boolean> = _reasoningEnabled

    private val _attachments = MutableStateFlow<List<Attachment>>(emptyList())
    val attachments: StateFlow<List<Attachment>> = _attachments

    private var client: HermeyClient? = null
    private var currentStreamId: String? = null
    private var sessionId: String? = null
    private var streamingThread: Thread? = null
    private val streamCancelled = AtomicBoolean(false)

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
                val json = c.getChatHistory(sessionId, 100)
                if (json.isNullOrBlank()) return@launch
                val array = JSONArray(json)
                val loaded = (0 until array.length()).map { i ->
                    val obj = array.getJSONObject(i)
                    UiMessage(
                        id = obj.optString("id", UUID.randomUUID().toString()),
                        role = when (obj.optString("role")) {
                            "assistant" -> MessageRole.Assistant
                            "system" -> MessageRole.System
                            else -> MessageRole.User
                        },
                        content = StringBuilder(obj.optString("content"))
                    )
                }
                _messages.value = loaded
            } catch (_: Exception) {
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
        _reasoningEnabled.value = !_reasoningEnabled.value
    }

    fun addAttachment(attachment: Attachment) {
        _attachments.value += attachment
    }

    fun removeAttachment(uri: String) {
        _attachments.value = _attachments.value.filter { it.uri != uri }
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
        _messages.value += userMessage
        _inputText.value = ""
        _attachments.value = emptyList()

        val assistantMessage = UiMessage(
            id = UUID.randomUUID().toString(),
            role = MessageRole.Assistant,
            isStreaming = true,
            reasoning = if (_reasoningEnabled.value) UiReasoning() else null
        )
        _messages.value += assistantMessage
        _isStreaming.value = true
        streamCancelled.set(false)

        viewModelScope.launch(Dispatchers.IO) {
            val c = ensureClient() ?: run {
                markStreamError("Server not configured")
                return@launch
            }
            try {
                uploadPendingAttachments(c, sid)

                val start = c.startChat(sid, text, "", model)
                val streamId = start?.streamID
                currentStreamId = streamId
                if (streamId.isNullOrBlank()) {
                    markStreamError("Failed to start stream")
                    return@launch
                }
                streamingThread = Thread({
                    try {
                        c.subscribeStream(streamId, makeListener { event ->
                            handleEvent(event, assistantMessage.id)
                        })
                    } catch (e: InterruptedException) {
                    } catch (e: Exception) {
                        if (!streamCancelled.get()) {
                            handleEvent(ChatEvent.Error(e.message ?: "Stream error"), assistantMessage.id)
                        }
                    }
                }, "hermdroid-sse-${streamId.takeLast(6)}").apply {
                    isDaemon = true
                    start()
                }
            } catch (e: Exception) {
                markStreamError(e.message ?: "Stream failed")
            }
        }
    }

    private fun uploadPendingAttachments(c: HermeyClient, sid: String) {
        val pending = _attachments.value
        if (pending.isEmpty()) return
        pending.forEach { attachment ->
            try {
                val input = getApplication<Application>().contentResolver
                    .openInputStream(Uri.parse(attachment.uri))
                    ?: return@forEach
                val bytes = input.use { it.readBytes() }
                c.uploadFile(sid, attachment.name, bytes, "")
            } catch (_: Exception) {
            }
        }
    }

    fun stop() {
        streamCancelled.set(true)
        val streamId = currentStreamId
        val thread = streamingThread
        viewModelScope.launch(Dispatchers.IO) {
            try {
                thread?.interrupt()
                if (!streamId.isNullOrBlank()) {
                    client?.cancelChat(streamId)
                }
            } catch (_: Exception) {
            } finally {
                _isStreaming.value = false
                finalizeAssistantMessage()
            }
        }
    }

    fun steer(text: String) {
        val sid = sessionId ?: return
        viewModelScope.launch(Dispatchers.IO) {
            try {
                client?.steerChat(sid, text)
            } catch (e: Exception) {
                _messages.value += UiMessage(
                    id = UUID.randomUUID().toString(),
                    role = MessageRole.System,
                    content = StringBuilder("Steer failed: ${e.message}")
                )
            }
        }
    }

    fun retry() {
        val lastUser = _messages.value.findLast { it.role == MessageRole.User } ?: return
        _inputText.value = lastUser.displayContent
        _messages.value = _messages.value.let { list ->
            val last = list.lastOrNull()
            if (last?.role == MessageRole.Assistant && last.isStreaming) list.dropLast(1) else list
        }
        send()
    }

    fun toggleReasoningExpanded(messageId: String) {
        _messages.value = _messages.value.map { msg ->
            if (msg.id == messageId && msg.reasoning != null) {
                msg.copy(reasoning = msg.reasoning.copy(expanded = !msg.reasoning.expanded))
            } else msg
        }
    }

    fun toggleToolExpanded(messageId: String, toolId: String) {
        _messages.value = _messages.value.map { msg ->
            if (msg.id == messageId) {
                msg.copy(toolCalls = msg.toolCalls.map { tool ->
                    if (tool.id == toolId) tool.copy(expanded = !tool.expanded) else tool
                }.toMutableList())
            } else msg
        }
    }

    private fun handleEvent(event: ChatEvent, assistantId: String) {
        _messages.value = _messages.value.map { msg ->
            if (msg.id != assistantId) return@map msg
            when (event) {
                is ChatEvent.Token -> msg.apply { content.append(event.text) }
                is ChatEvent.Reasoning -> msg.apply {
                    reasoning?.text?.append(event.text)
                }
                is ChatEvent.ToolCall -> msg.apply { toolCalls.add(parseToolCall(event.json)) }
                is ChatEvent.ToolResult -> msg.apply { attachToolResult(event.json) }
                is ChatEvent.Error -> msg.copy(streamError = event.message, isStreaming = false)
                is ChatEvent.StreamEnd, ChatEvent.Cancel -> msg.copy(isStreaming = false)
            }
        }
        if (event is ChatEvent.StreamEnd || event is ChatEvent.Error || event is ChatEvent.Cancel) {
            _isStreaming.value = false
            finalizeAssistantMessage()
        }
    }

    private fun finalizeAssistantMessage() {
        _messages.value = _messages.value.map { msg ->
            if (msg.role == MessageRole.Assistant && msg.isStreaming) {
                msg.copy(isStreaming = false)
            } else msg
        }
    }

    private fun markStreamError(message: String) {
        _isStreaming.value = false
        _messages.value = _messages.value.let { list ->
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

    private fun makeListener(onEvent: (ChatEvent) -> Unit): EventListenerProxy {
        return object : EventListenerProxy {
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
