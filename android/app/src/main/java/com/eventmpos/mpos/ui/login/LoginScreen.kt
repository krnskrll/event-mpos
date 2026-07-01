package com.eventmpos.mpos.ui.login

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Spa
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.eventmpos.mpos.ui.components.BrandButton
import com.eventmpos.mpos.ui.theme.Brand

@Composable
fun LoginScreen(
    onLoggedIn: () -> Unit,
    viewModel: LoginViewModel = hiltViewModel(),
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    LaunchedEffect(state.loggedIn) { if (state.loggedIn) onLoggedIn() }

    Column(
        modifier = Modifier.fillMaxSize().background(Color.White).padding(horizontal = 24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Box(
            modifier = Modifier.size(80.dp).clip(RoundedCornerShape(20.dp)).background(Brand.Navy),
            contentAlignment = Alignment.Center,
        ) { Icon(Icons.Outlined.Spa, contentDescription = null, tint = Brand.Teal, modifier = Modifier.size(40.dp)) }
        Text("BEAUTY POS", color = Brand.Teal, fontSize = 12.sp, fontWeight = FontWeight.Bold, letterSpacing = 3.sp, modifier = Modifier.padding(top = 14.dp))

        Text("Welcome Back", color = Brand.Heading, fontSize = 26.sp, fontWeight = FontWeight.Bold, modifier = Modifier.padding(top = 24.dp))
        Text("Sign in to continue", color = Brand.TextFaint, fontSize = 15.sp, modifier = Modifier.padding(top = 2.dp, bottom = 24.dp))

        Field(label = "Email", value = state.email, onChange = viewModel::onEmailChange, keyboard = KeyboardType.Email)
        Field(label = "Password", value = state.password, onChange = viewModel::onPasswordChange, keyboard = KeyboardType.Password, password = true)

        if (state.error != null) {
            Text(state.error!!, color = Brand.Danger, fontSize = 13.sp, modifier = Modifier.fillMaxWidth().padding(top = 4.dp))
        }

        BrandButton(
            text = "Sign In",
            onClick = viewModel::login,
            loading = state.loading,
            modifier = Modifier.fillMaxWidth().padding(top = 20.dp),
        )
        Text("Powered by event-mpos", color = Brand.TextGhost, fontSize = 12.sp, modifier = Modifier.padding(top = 28.dp))
    }
}

@Composable
private fun Field(
    label: String,
    value: String,
    onChange: (String) -> Unit,
    keyboard: KeyboardType,
    password: Boolean = false,
) {
    Column(modifier = Modifier.fillMaxWidth().padding(top = 12.dp)) {
        Text(label, color = Color(0xFF374151), fontSize = 14.sp, fontWeight = FontWeight.Medium, modifier = Modifier.padding(bottom = 6.dp))
        OutlinedTextField(
            value = value,
            onValueChange = onChange,
            singleLine = true,
            visualTransformation = if (password) PasswordVisualTransformation() else androidx.compose.ui.text.input.VisualTransformation.None,
            keyboardOptions = KeyboardOptions(
                keyboardType = keyboard,
                autoCorrect = !password,
                capitalization = KeyboardCapitalization.None,
            ),
            shape = RoundedCornerShape(12.dp),
            colors = OutlinedTextFieldDefaults.colors(
                focusedBorderColor = Brand.Navy,
                unfocusedBorderColor = Brand.FieldBorder,
                focusedContainerColor = Brand.FieldBg,
                unfocusedContainerColor = Brand.FieldBg,
            ),
            modifier = Modifier.fillMaxWidth(),
        )
    }
}
