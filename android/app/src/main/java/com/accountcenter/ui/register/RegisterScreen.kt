package com.accountcenter.ui.register

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.accountcenter.ui.components.GradientButton
import com.accountcenter.ui.theme.*

@Composable
fun RegisterScreen(
    viewModel: RegisterViewModel = hiltViewModel(),
    onNavigateBack: () -> Unit,
    onRegisterSuccess: () -> Unit
) {
    val uiState = viewModel.uiState.collectAsState().value
    Box(modifier = Modifier.fillMaxSize().background(BgPrimary)) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState())
                .padding(32.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(24.dp))
            Column(modifier = Modifier.padding(bottom = 32.dp), horizontalAlignment = Alignment.CenterHorizontally) {
                Icon(
                    imageVector = Icons.Default.Person,
                    contentDescription = null,
                    modifier = Modifier.size(56.dp),
                    tint = BrandPrimary
                )
                Spacer(modifier = Modifier.height(12.dp))
                Text("创建账户", style = MaterialTheme.typography.headlineLarge, color = TextPrimary)
                Text("注册新账户以使用完整功能", style = MaterialTheme.typography.bodyMedium, color = TextSecondary)
            }

            OutlinedTextField(
                value = uiState.phoneNumber,
                onValueChange = viewModel::onPhoneNumberChange,
                label = { Text("手机号") },
                singleLine = true,
                keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Phone),
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(16.dp))

            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    value = uiState.verificationCode,
                    onValueChange = viewModel::onVerificationCodeChange,
                    label = { Text("验证码") },
                    singleLine = true,
                    keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                    modifier = Modifier.weight(1f),
                    colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
                )
                val btnText = if (uiState.countdownSeconds > 0) "${uiState.countdownSeconds}s" else "发送验证码"
                OutlinedButton(
                    onClick = viewModel::sendVerificationCode,
                    enabled = !uiState.isLoading && uiState.countdownSeconds == 0 && uiState.phoneNumber.isNotEmpty()
                ) { Text(btnText, color = BrandSecondary) }
            }
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = uiState.accountId,
                onValueChange = viewModel::onAccountIdChange,
                label = { Text("账户ID") },
                singleLine = true,
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = uiState.password,
                onValueChange = viewModel::onPasswordChange,
                label = { Text("密码") },
                singleLine = true,
                visualTransformation = PasswordVisualTransformation(),
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = uiState.confirmPassword,
                onValueChange = viewModel::onConfirmPasswordChange,
                label = { Text("确认密码") },
                singleLine = true,
                visualTransformation = PasswordVisualTransformation(),
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = uiState.referralCode,
                onValueChange = viewModel::onReferralCodeChange,
                label = { Text("推荐码（可选）") },
                singleLine = true,
                modifier = Modifier.fillMaxWidth(),
                colors = OutlinedTextFieldDefaults.colors(focusedBorderColor = BrandSecondary, unfocusedBorderColor = Divider, focusedLabelColor = BrandSecondary, unfocusedLabelColor = TextSecondary, cursorColor = BrandSecondary)
            )
            Spacer(modifier = Modifier.height(8.dp))

            Row(verticalAlignment = Alignment.CenterVertically) {
                Checkbox(
                    checked = uiState.agreeToTerms,
                    onCheckedChange = viewModel::onAgreeToTermsChange,
                    colors = CheckboxDefaults.colors(checkedColor = BrandPrimary)
                )
                Text("我已阅读并同意服务条款和隐私政策", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
            }

            uiState.errorMessage?.let {
                Text(it, color = Danger, style = MaterialTheme.typography.bodySmall)
            }
            Spacer(modifier = Modifier.height(16.dp))

            GradientButton(
                text = "注册",
                onClick = { viewModel.register(onRegisterSuccess) },
                enabled = !uiState.isLoading
            )

            Spacer(modifier = Modifier.height(16.dp))
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Center) {
                Text("已有账号？", style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                TextButton(onClick = onNavigateBack) {
                    Text("立即登录", color = BrandSecondary)
                }
            }
            Spacer(modifier = Modifier.height(24.dp))
        }

        if (uiState.isLoading) {
            CircularProgressIndicator(modifier = Modifier.align(Alignment.Center), color = BrandSecondary)
        }
    }
}
