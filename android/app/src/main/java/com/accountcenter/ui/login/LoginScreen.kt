package com.accountcenter.ui.login

import android.view.WindowManager
import androidx.activity.compose.LocalActivity
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Person
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.accountcenter.R
import com.accountcenter.ui.components.GradientButton
import com.accountcenter.ui.theme.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LoginScreen(
    viewModel: LoginViewModel = hiltViewModel(),
    onNavigateToRegister: () -> Unit,
    onLoginSuccess: () -> Unit
) {
    val activity = LocalActivity.current
    DisposableEffect(activity) {
        activity?.window?.addFlags(WindowManager.LayoutParams.FLAG_SECURE)
        onDispose {
            activity?.window?.clearFlags(WindowManager.LayoutParams.FLAG_SECURE)
        }
    }

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
                tint = BrandPrimary
            )
            Spacer(modifier = Modifier.height(12.dp))
            Text(
                stringResource(R.string.login_title),
                style = MaterialTheme.typography.headlineLarge,
                color = TextPrimary
            )
            Text(
                stringResource(R.string.login_subtitle),
                style = MaterialTheme.typography.bodyMedium,
                color = TextSecondary
            )
            Spacer(modifier = Modifier.height(48.dp))

            OutlinedTextField(
                value = uiState.phoneNumber,
                onValueChange = viewModel::onPhoneNumberChange,
                label = { Text(stringResource(R.string.login_phone_label)) },
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
            Spacer(modifier = Modifier.height(16.dp))

            SingleChoiceSegmentedButtonRow(modifier = Modifier.fillMaxWidth()) {
                SegmentedButton(
                    selected = uiState.loginMode == LoginMode.PASSWORD,
                    onClick = { viewModel.onLoginModeChange(LoginMode.PASSWORD) },
                    shape = SegmentedButtonDefaults.itemShape(index = 0, count = 2)
                ) {
                    Text(stringResource(R.string.login_password_mode), color = TextPrimary)
                }
                SegmentedButton(
                    selected = uiState.loginMode == LoginMode.VERIFICATION_CODE,
                    onClick = { viewModel.onLoginModeChange(LoginMode.VERIFICATION_CODE) },
                    shape = SegmentedButtonDefaults.itemShape(index = 1, count = 2)
                ) {
                    Text(stringResource(R.string.login_code_mode), color = TextPrimary)
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            if (uiState.loginMode == LoginMode.PASSWORD) {
                OutlinedTextField(
                    value = uiState.password,
                    onValueChange = viewModel::onPasswordChange,
                    label = { Text(stringResource(R.string.login_password_placeholder)) },
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
                        label = { Text(stringResource(R.string.register_code_placeholder)) },
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
                    val btnText = if (uiState.countdownSeconds > 0) "${uiState.countdownSeconds}s" else stringResource(R.string.send_code)
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
            Spacer(modifier = Modifier.height(16.dp))

            GradientButton(
                text = stringResource(R.string.login_button),
                onClick = { viewModel.login(onLoginSuccess) },
                enabled = !uiState.isLoading
            )

            Spacer(modifier = Modifier.height(24.dp))
            Row(horizontalArrangement = Arrangement.Center, modifier = Modifier.fillMaxWidth()) {
                Text(stringResource(R.string.login_no_account), style = MaterialTheme.typography.bodySmall, color = TextSecondary)
                TextButton(onClick = onNavigateToRegister) {
                    Text(stringResource(R.string.login_register), color = BrandSecondary)
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
