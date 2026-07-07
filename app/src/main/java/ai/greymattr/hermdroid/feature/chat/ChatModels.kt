package ai.greymattr.hermdroid.feature.chat

data class Attachment(
    val uri: String,
    val name: String,
    val mimeType: String,
    val sizeBytes: Long,
    val isImage: Boolean
)

data class UiToolCall(
    val id: String,
    val name: String,
    val input: String,
    val result: String? = null,
    val status: ToolStatus = ToolStatus.Pending,
    val expanded: Boolean = false
)

enum class ToolStatus { Pending, Running, Success, Error }

data class UiReasoning(
    val text: StringBuilder = StringBuilder(),
    val expanded: Boolean = false
)

data class UiMessage(
    val id: String,
    val role: MessageRole,
    val content: StringBuilder = StringBuilder(),
    val reasoning: UiReasoning? = null,
    val toolCalls: MutableList<UiToolCall> = mutableListOf(),
    val attachments: List<Attachment> = emptyList(),
    val isStreaming: Boolean = false,
    val streamError: String? = null,
    val timestamp: Long = System.currentTimeMillis()
) {
    val hasReasoning: Boolean get() = reasoning != null && reasoning.text.isNotBlank()
    val displayContent: String get() = content.toString()
}

enum class MessageRole { User, Assistant, System }

sealed class ChatEvent {
    data class Token(val text: String) : ChatEvent()
    data class ToolCall(val json: String) : ChatEvent()
    data class ToolResult(val json: String) : ChatEvent()
    data class Reasoning(val text: String) : ChatEvent()
    object StreamEnd : ChatEvent()
    data class Error(val message: String) : ChatEvent()
    object Cancel : ChatEvent()
}
