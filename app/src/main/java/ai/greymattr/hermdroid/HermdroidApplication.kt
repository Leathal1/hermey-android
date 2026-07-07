package ai.greymattr.hermdroid

import android.app.Application
import ai.greymattr.hermdroid.data.local.SecureSettings
import ai.greymattr.hermdroid.data.local.createCacheDirectory

class HermdroidApplication : Application() {

    override fun onCreate() {
        super.onCreate()
        createCacheDirectory()
        SecureSettings.init(this)
    }
}
