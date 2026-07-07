package ai.greymattr.hermdroid.feature.settings

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import ai.greymattr.hermdroid.data.local.SecureSettings

class SettingsViewModel(application: Application) : AndroidViewModel(application) {
    var serverUrl: String
        get() = SecureSettings.serverUrl
        set(value) { SecureSettings.serverUrl = value }

    var authToken: String
        get() = SecureSettings.authToken
        set(value) { SecureSettings.authToken = value }

    var defaultModel: String
        get() = SecureSettings.defaultModel
        set(value) { SecureSettings.defaultModel = value }

    var themeMode: String
        get() = SecureSettings.themeMode
        set(value) { SecureSettings.themeMode = value }
}
