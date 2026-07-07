package ai.greymattr.hermdroid.data.repository

import ai.greymattr.hermdroid.core.cache.CachedMessage
import ai.greymattr.hermdroid.core.cache.CachedSession
import ai.greymattr.hermdroid.core.cache.HermdroidCacheLib
import ai.greymattr.hermdroid.core.cache.parseMessagesJson
import ai.greymattr.hermdroid.core.cache.parseSessionsJson
import android.content.Context
import android.util.Log
import java.io.File

private const val TAG = "OfflineRepository"
private const val CACHE_DB = "hermdroid_cache.db"
private const val DEFAULT_MAX_MESSAGES = 10_000
private const val DEFAULT_MAX_BYTES = 50L * 1024 * 1024

/**
 * Read-only offline cache repository.
 *
 * Exposes data persisted by the native Go bbolt cache when the device is offline.
 * All methods are safe to call when the library is unavailable; they return empty results.
 */
class OfflineRepository(context: Context) {

    private val dbPath = File(context.cacheDir, CACHE_DB).absolutePath
    private val handle: Int by lazy {
        if (HermdroidCacheLib.available) {
            HermdroidCacheLib.open(dbPath, DEFAULT_MAX_MESSAGES, DEFAULT_MAX_BYTES)
        } else {
            -1
        }
    }

    val isAvailable: Boolean get() = HermdroidCacheLib.available && handle >= 0

    fun sessions(): List<CachedSession> {
        if (!isAvailable) return emptyList()
        return try {
            parseSessionsJson(HermdroidCacheLib.listSessionsJson(handle))
        } catch (e: Exception) {
            Log.e(TAG, "Failed to list sessions", e)
            emptyList()
        }
    }

    fun messages(sessionId: String): List<CachedMessage> {
        if (!isAvailable) return emptyList()
        return try {
            parseMessagesJson(HermdroidCacheLib.getMessagesJson(handle, sessionId))
        } catch (e: Exception) {
            Log.e(TAG, "Failed to load messages", e)
            emptyList()
        }
    }

    fun putSession(session: CachedSession) {
        if (!isAvailable) return
        try {
            HermdroidCacheLib.putSession(handle, session.id, session.title, session.lastMessageAtMs, session.messageCount)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to put session", e)
        }
    }

    fun putMessage(message: CachedMessage) {
        if (!isAvailable) return
        try {
            HermdroidCacheLib.putMessage(handle, message.id, message.sessionId, message.role, message.content, message.timestampMs)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to put message", e)
        }
    }

    fun evict() {
        if (!isAvailable) return
        try {
            HermdroidCacheLib.evict(handle)
        } catch (e: Exception) {
            Log.e(TAG, "Failed to evict cache", e)
        }
    }

    fun close() {
        if (handle >= 0) {
            try {
                HermdroidCacheLib.close(handle)
            } catch (e: Exception) {
                Log.e(TAG, "Failed to close cache", e)
            }
        }
    }
}
