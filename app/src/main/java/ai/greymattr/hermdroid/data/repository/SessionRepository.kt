package ai.greymattr.hermdroid.data.repository

import ai.greymattr.hermdroid.data.local.SecureSettings
import android.app.Application
import core.HermeyClient
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONArray
import org.json.JSONObject
import java.text.SimpleDateFormat
import java.util.Locale
import java.util.TimeZone

sealed class SessionResult<out T> {
    data class Success<T>(val value: T) : SessionResult<T>()
    data class Error(val message: String, val offline: Boolean = false) : SessionResult<Nothing>()
}

class SessionRepository(application: Application) {

    private var client: HermeyClient? = null

    private fun clientOrError(): HermeyClient? {
        if (client != null) return client
        val url = SecureSettings.serverUrl.takeIf { it.isNotBlank() } ?: return null
        return try {
            HermeyClient(url).also { client = it }
        } catch (e: Exception) {
            null
        }
    }

    private fun <T> resultFrom(block: () -> T): SessionResult<T> =
        try {
            SessionResult.Success(block())
        } catch (e: Exception) {
            val offline = e.message?.contains("Unable to resolve host", ignoreCase = true) == true ||
                    e.message?.contains("Failed to connect", ignoreCase = true) == true ||
                    e.message?.contains("timeout", ignoreCase = true) == true
            SessionResult.Error(e.message ?: "Unknown error", offline)
        }

    private fun isoToMillis(iso: String?): Long {
        if (iso.isNullOrBlank()) return 0L
        return try {
            val fmt = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss.SSS'Z'", Locale.US)
            fmt.timeZone = TimeZone.getTimeZone("UTC")
            fmt.parse(iso)?.time ?: 0L
        } catch (_: Exception) {
            0L
        }
    }

    private fun parseSessions(json: String?): List<UiSession> {
        if (json.isNullOrBlank()) return emptyList()
        val array = JSONArray(json)
        return (0 until array.length()).map { i ->
            val obj = array.getJSONObject(i)
            obj.toUiSession()
        }
    }

    private fun JSONObject.toUiSession(): UiSession = UiSession(
        id = optString("id", ""),
        title = optString("title", optString("id", "Untitled")),
        model = optString("model", ""),
        profile = optString("profile", ""),
        workspace = optString("workspace", ""),
        messageCount = optInt("message_count", optInt("messageCount", 0)),
        pinned = optBoolean("pinned", false),
        archived = optBoolean("archived", false),
        lastMessageAtMs = isoToMillis(optString("last_message_at", optString("lastMessageAt", ""))),
        projectId = optString("project_id", optString("projectID", ""))
    )

    suspend fun sessions(): SessionResult<List<UiSession>> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom { parseSessions(c.listSessions()) }
    }

    suspend fun search(query: String, limit: Int = 50): SessionResult<List<UiSession>> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom { parseSessions(c.searchSessionsJson(query, limit.toLong())) }
    }

    suspend fun delete(sessionId: String): SessionResult<Unit> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom { c.deleteSession(sessionId) }
    }

    suspend fun rename(sessionId: String, title: String): SessionResult<Unit> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom { c.renameSession(sessionId, title) }
    }

    suspend fun pin(sessionId: String, pinned: Boolean): SessionResult<Unit> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom { c.pinSession(sessionId, pinned) }
    }

    suspend fun archive(sessionId: String, archived: Boolean): SessionResult<Unit> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom { c.archiveSession(sessionId, archived) }
    }

    suspend fun branch(sessionId: String, title: String): SessionResult<String> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom { c.forkSessionJson(sessionId, title) ?: "" }
    }

    suspend fun moveToProject(sessionId: String, projectId: String): SessionResult<Unit> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom { c.moveSession(sessionId, projectId) }
    }

    suspend fun truncate(sessionId: String, messageId: String = "", keepCount: Int = 0): SessionResult<Unit> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom { c.truncateSession(sessionId, messageId, keepCount.toLong()) }
    }

    suspend fun newSession(workspace: String, model: String, profile: String, title: String = ""): SessionResult<UiSession> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom {
            val json = c.newSessionJson(workspace, model, "", profile, title)
            val obj = JSONObject(json)
            obj.toUiSession().copy(
                title = obj.optString("title", title.ifBlank { "New chat" })
            )
        }
    }

    suspend fun projects(): SessionResult<List<UiProject>> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom {
            if (c.listProjectsJson().isNullOrBlank()) emptyList()
            else JSONArray(c.listProjectsJson()).let { array ->
                (0 until array.length()).map { i ->
                    val obj = array.getJSONObject(i)
                    UiProject(
                        id = obj.optString("id", ""),
                        name = obj.optString("name", obj.optString("id", ""))
                    )
                }
            }
        }
    }

    suspend fun models(): SessionResult<List<ModelOption>> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom {
            if (c.listModelsJson().isNullOrBlank()) emptyList()
            else JSONArray(c.listModelsJson()).let { array ->
                (0 until array.length()).map { i ->
                    val obj = array.getJSONObject(i)
                    ModelOption(
                        id = obj.optString("id", ""),
                        name = obj.optString("name", obj.optString("id", "")),
                        provider = obj.optString("provider", "")
                    )
                }
            }
        }
    }

    suspend fun profiles(): SessionResult<List<ProfileOption>> = withContext(Dispatchers.IO) {
        val c = clientOrError()
            ?: return@withContext SessionResult.Error("Server URL not configured", offline = false)
        resultFrom {
            if (c.listProfilesJson().isNullOrBlank()) emptyList()
            else JSONArray(c.listProfilesJson()).let { array ->
                (0 until array.length()).map { i ->
                    val obj = array.getJSONObject(i)
                    ProfileOption(
                        id = obj.optString("name", obj.optString("id", "")),
                        description = obj.optString("description", ""),
                        isActive = obj.optBoolean("active", obj.optBoolean("is_active", false))
                    )
                }
            }
        }
    }
}

data class UiSession(
    val id: String,
    val title: String,
    val model: String = "",
    val profile: String = "",
    val workspace: String = "",
    val messageCount: Int = 0,
    val pinned: Boolean = false,
    val archived: Boolean = false,
    val lastMessageAtMs: Long = 0L,
    val projectId: String = ""
) {
    val isEmpty: Boolean get() = messageCount == 0
}

data class UiProject(val id: String, val name: String)

data class ModelOption(val id: String, val name: String, val provider: String)

data class ProfileOption(val id: String, val description: String, val isActive: Boolean)
