package ai.greymattr.hermdroid.core.cache

import org.json.JSONArray
import org.json.JSONObject

/**
 * Domain models for the offline read-only cache.
 */
data class CachedSession(
    val id: String,
    val title: String,
    val lastMessageAtMs: Long,
    val messageCount: Int
)

data class CachedMessage(
    val id: String,
    val sessionId: String,
    val role: String,
    val content: String,
    val timestampMs: Long
)

fun parseSessionsJson(json: String): List<CachedSession> {
    if (json.isBlank()) return emptyList()
    val arr = JSONArray(json)
    return (0 until arr.length()).map { i ->
        val o = arr.getJSONObject(i)
        CachedSession(
            id = o.getString("id"),
            title = o.optString("title", ""),
            lastMessageAtMs = o.optLong("lastMessageAt", 0L),
            messageCount = o.optInt("messageCount", 0)
        )
    }
}

fun parseMessagesJson(json: String): List<CachedMessage> {
    if (json.isBlank()) return emptyList()
    val arr = JSONArray(json)
    return (0 until arr.length()).map { i ->
        val o = arr.getJSONObject(i)
        CachedMessage(
            id = o.getString("id"),
            sessionId = o.optString("sessionId", ""),
            role = o.optString("role", ""),
            content = o.optString("content", ""),
            timestampMs = o.optLong("timestamp", 0L)
        )
    }
}

fun CachedMessage.toJson(): String = JSONObject().apply {
    put("id", id)
    put("sessionId", sessionId)
    put("role", role)
    put("content", content)
    put("timestamp", timestampMs)
}.toString()
