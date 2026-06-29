package com.eventmpos.mpos

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import com.eventmpos.mpos.domain.auth.AuthRepository
import com.eventmpos.mpos.ui.AppNavHost
import com.eventmpos.mpos.ui.theme.EventMposTheme
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    @Inject
    lateinit var authRepository: AuthRepository

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        // Decide the start screen up front from stored tokens: a returning BA skips login.
        val startLoggedIn = authRepository.isLoggedIn()
        setContent {
            EventMposTheme {
                Surface(modifier = Modifier.fillMaxSize(), color = MaterialTheme.colorScheme.background) {
                    AppNavHost(startLoggedIn = startLoggedIn)
                }
            }
        }
    }
}
