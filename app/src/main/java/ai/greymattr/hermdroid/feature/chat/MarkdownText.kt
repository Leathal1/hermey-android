package ai.greymattr.hermdroid.feature.chat

import android.content.Context
import android.text.method.LinkMovementMethod
import android.widget.TextView
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.viewinterop.AndroidView
import androidx.compose.material3.MaterialTheme
import io.noties.markwon.AbstractMarkwonPlugin
import io.noties.markwon.Markwon
import io.noties.markwon.ext.strikethrough.StrikethroughPlugin
import io.noties.markwon.ext.tables.TablePlugin
import io.noties.markwon.html.HtmlPlugin
import io.noties.markwon.image.coil.CoilImagesPlugin
import io.noties.markwon.linkify.LinkifyPlugin

/**
 * Renders Markdown text using Markwon inside an AndroidView. This keeps the
 * rich text, code blocks, tables, links, and inline images intact while
 * streaming incremental content.
 */
@Composable
fun MarkdownText(
    markdown: String,
    modifier: Modifier = Modifier,
) {
    val colorScheme = MaterialTheme.colorScheme
    val textColor = remember(colorScheme) { colorScheme.onSurface.toArgb() }
    val codeBackground = remember(colorScheme) { colorScheme.surfaceVariant.toArgb() }
    val context = androidx.compose.ui.platform.LocalContext.current
    val markwon = remember(context, textColor, codeBackground) {
        createMarkwon(context, textColor, codeBackground)
    }

    AndroidView(
        factory = { ctx ->
            TextView(ctx).apply {
                setTextIsSelectable(true)
                movementMethod = LinkMovementMethod.getInstance()
                setLineSpacing(0f, 1.15f)
            }
        },
        update = { textView ->
            markwon.setMarkdown(textView, markdown)
        },
        modifier = modifier.fillMaxWidth()
    )
}

private fun createMarkwon(context: Context, textColor: Int, codeBackground: Int): Markwon {
    return Markwon.builder(context)
        .usePlugin(object : AbstractMarkwonPlugin() {
            override fun configureTheme(builder: io.noties.markwon.core.MarkwonTheme.Builder) {
                builder
                    .codeTextColor(textColor)
                    .codeBackgroundColor(codeBackground)
                    .blockQuoteColor(textColor and 0x80FFFFFF.toInt())
            }
        })
        .usePlugin(HtmlPlugin.create())
        .usePlugin(LinkifyPlugin.create())
        .usePlugin(StrikethroughPlugin.create())
        .usePlugin(TablePlugin.create(context))
        .usePlugin(CoilImagesPlugin.create(context))
        .build()
}
