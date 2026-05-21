import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.unit.em

object DesignTokens {
    object Colors {
        val BgPrimary = Color(0xFF0D1117)
        val BgSecondary = Color(0xFF161B22)
        val BgCard = Color(0xFF1C2333)
        val BgInput = Color(0xFF161B22)

        val BrandPrimary = Color(0xFF6C63FF)
        val BrandSecondary = Color(0xFF00D4FF)

        val TextPrimary = Color(0xDEFFFFFF)
        val TextSecondary = Color(0xFF8B949E)
        val TextDisabled = Color(0xFF484F58)

        val BorderDefault = Color(0xFF30363D)
        val BorderMuted = Color(0xFF21262D)

        val Success = Color(0xFF2ED573)
        val Warning = Color(0xFFFFA502)
        val Danger = Color(0xFFFF4757)
        val Info = Color(0xFF00D4FF)

        val TierFree = Color(0xFF8B949E)
        val TierBasic = Color(0xFF6C63FF)
        val TierPremium = Color(0xFFFF9800)
        val TierEnterprise = Color(0xFF7B1FA2)
    }

    object Typography {
        val FontSizeXS = 12.sp
        val FontSizeSM = 14.sp
        val FontSizeMD = 16.sp
        val FontSizeLG = 18.sp
        val FontSizeXL = 20.sp
        val FontSize2XL = 24.sp
        val FontSize3XL = 32.sp
        val FontSize4XL = 48.sp

        val LineHeightTight = 1.25.em
        val LineHeightNormal = 1.5.em
        val LineHeightRelaxed = 1.75.em
    }

    object Spacing {
        val XS = 4.dp
        val SM = 8.dp
        val MD = 16.dp
        val LG = 24.dp
        val XL = 32.dp
        val XXL = 48.dp
        val XXXL = 64.dp
    }

    object BorderRadius {
        val SM = 4.dp
        val MD = 8.dp
        val LG = 12.dp
        val XL = 16.dp
        val XXL = 24.dp
    }

    object Motion {
        const val DurationFast = 150
        const val DurationNormal = 250
        const val DurationSlow = 350
    }
}
