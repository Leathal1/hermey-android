package ai.greymattr.hermdroid.data.local

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey

object SecureSettings {
    private const val FILE = "hermdroid_secure_settings"

    @Volatile
    private var prefs: SharedPreferences? = null

    fun init(context: Context) {
        if (prefs != null) return
        val masterKey = MasterKey.Builder(context)
            .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
            .build()
        prefs = EncryptedSharedPreferences.create(
            context,
            FILE,
            masterKey,
            EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
            EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
        )
    }

    private fun requirePrefs(): SharedPreferences {
        return prefs ?: throw IllegalStateException("SecureSettings not initialized")
    }

    var serverUrl: String
        get() = requirePrefs().getString(KEY_SERVER_URL, "") ?: ""
        set(value) = requirePrefs().edit().putString(KEY_SERVER_URL, value).apply()

    var authToken: String
        get() = requirePrefs().getString(KEY_AUTH_TOKEN, "") ?: ""
        set(value) = requirePrefs().edit().putString(KEY_AUTH_TOKEN, value).apply()

    var defaultModel: String
        get() = requirePrefs().getString(KEY_DEFAULT_MODEL, "default") ?: "default"
        set(value) = requirePrefs().edit().putString(KEY_DEFAULT_MODEL, value).apply()

    var themeMode: String
        get() = requirePrefs().getString(KEY_THEME_MODE, "system") ?: "system"
        set(value) = requirePrefs().edit().putString(KEY_THEME_MODE, value).apply()

    fun clear() {
        requirePrefs().edit().clear().apply()
    }

    private const val KEY_SERVER_URL = "server_url"
    private const val KEY_AUTH_TOKEN = "auth_token"
    private const val KEY_DEFAULT_MODEL = "default_model"
    private const val KEY_THEME_MODE = "theme_mode"
}

fun Context.createCacheDirectory(): java.io.File {
    return cacheDir.also { it.mkdirs() }
}
