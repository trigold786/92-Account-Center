package com.accountcenter.ui.login

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.accountcenter.ui.components.GradientButton
import com.accountcenter.ui.theme.*

@Composable
fun LoginScreen(
    viewModel: LoginViewModel = hiltViewModel(),
    onNavigateToRegister: () -> Unit,
    onLoginSuccess: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsState()

    Box(
        modifier = Modifier.fillMaxSize().background(BgPrimary)
    ) {
        Column(
            modifier = Modifier.fillMaxSize().padding(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Icon(
                imageVector = Icons.Default.Person,
                contentDescription = null,
                modifier = Modifier.size(56.dp),
                tint = Brush.horizontalGradient(listOf(BrandPrimary, BrandSecondary))
            )
            Spacer().height(12.dp)
            Text(
                "账户中心",
                style = MaterialTheme.typography.headlineLarge,
                color = TextPrimary
            )
            Text(
                "登录您的账户",
                style = MaterialTheme.typography.bodyMedium,
                color = TextSecondary
            )
            Spacer().height(48.dp)

            OutlinedTextField(
                value = uiState.phoneNumber,
                onValueChange = viewModel::onPhoneNumberChange,
                label = { Text("手机号") },
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedBorderColor = BrandSecondary,
                    unfocusedBorderColor = Divider,
                    focusedLabelColor = BrandSecondary,
                    unfocusedLabelColor = TextSecondary,
                    cursorColor = BrandSecondary
                )
            )
            Spacer().height(16.dp)

            SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                SegmentedButton(
                    selected = uiState.loginMode == LoginMode.PASSWORD,
                    onClick = { viewModel.onLoginModeChange(LoginMode.PASSWORD) },
                    shape = SegmentedButtonDefaults.itemShape(index = 0, count = 2)
                ) {
                    Text("密码登录", color = TextPrimary)
                }
                SegmentedButton(
                    selected = uiState.loginMode == LoginMode.VERIFICATION_CODE,
                    onClick = { viewModel.onLoginModeChange(LoginMode.VERIFICATION_CODE) },
                    shape = SegmentedButtonDefaults.itemShape(index = 1, count = 2)
                ) {
                    Text("验证码登录", color = TextPrimary)
                }
            }

            Spacer().height(16.dp)

            if (uiState.loginMode == LoginMode.PASSWORD) {
                OutlinedTextField(
                    value = uiState.password,
                    onValueChange = viewModel::onPasswordChange,
                    label = { Text("密码") },
                    singleLine = true,
                    visualTransformation = PasswordVisualTransformation(),
                    modifier = Modifier.fillMaxWidth(),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = BrandSecondary,
                        unfocusedBorderColor = Divider,
                        focusedLabelColor = BrandSecondary,
                        unfocusedLabelColor = TextSecondary,
                        cursorColor = BrandSecondary
                    )
                )
            } else {
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedTextField(
                        value = uiState.verificationCode,
                        onValueChange = viewModel::onVerificationCodeChange,
                        label = { Text("验证码") },
                        singleLine = true,
                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                        modifier = Modifier.weight(1f),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = BrandSecondary,
                            unfocusedBorderColor = Divider,
                            focusedLabelColor = BrandSecondary,
                            unfocusedLabelColor = TextSecondary,
                            cursorColor = BrandSecondary
                        )
                    )
                    val btnText = if (uiState.countdownSeconds > 0) "${uiState.countdownSeconds}s" else "发送验证码"
                    OutlinedButton(
                        onClick = viewModel::sendVerificationCode,
                        enabled = !uiState.isLoading && uiState.countdownSeconds == 0 && uiState.phoneNumber.isNotEmpty()
                    ) {
                        Text(btnText, color = BrandSecondary)
                    }
                }
            }

            uiState.errorMessage?.let {
                Text(it, color = Danger, style = MaterialTheme.typography.bodySmall)
            }
            Spacer().height(16.dp)

            GradientButton(
                text = "登录",
                onClick = { viewModel.login(onLoginSuccess) },
                enabled = !uiState.isLoading
            )

            Spacer().height(24.dp)
            Row(horizontalArrangement = Arrangement.Center, modifier = Modifier.fillMaxWidth()) {
                Text("还没有账号？", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                TextButton(onClick = onNavigateToRegister) {
                    Text("立即注册", color = BrandSecondary)
                }
            }
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(
                modifier = Modifier.align(Alignment.Center),
                color = BrandSecondary
            )
        }
    }
}
