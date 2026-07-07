package ai.greymattr.hermdroid.core.cache

import android.util.Log

/**
 * JNI bridge to the Go bbolt cache shared library.
 *
 * The native library is built from `core/cachebridge/` using:
 *   gomobile bind -target=android -o app/libs/cachebridge.aar ./core/cachebridge
 *
 * On host JVM tests the library can be loaded from `core/cachebridge/libhermdroidcache.so`
 * built with `go build -buildmode=c-shared`.
 */
object HermdroidCacheLib {

    private const val TAG = "HermdroidCache"
    private var loaded = false

    init {
        try {
            System.loadLibrary("hermdroidcache")
            loaded = true
        } catch (e: UnsatisfiedLinkError) {
            Log.w(TAG, "Native cache library not available: ${e.message}")
        }
    }

    val available: Boolean get() = loaded

    external fun open(path: String, maxMessages: Int, maxBytes: Long): Int
    external fun close(handle: Int)
    external fun putSession(handle: Int, id: String, title: String, lastMessageAtMs: Long, count: Int): Int
    external fun listSessionsJson(handle: Int): String
    external fun getMessagesJson(handle: Int, sessionId: String): String
    external fun putMessage(handle: Int, id: String, sessionId: String, role: String, content: String, timestampMs: Long): Int
    external fun deleteSession(handle: Int, sessionId: String): Int
    external fun evict(handle: Int): Int
}
