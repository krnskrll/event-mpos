package com.eventmpos.mpos.ui.cart

import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.eventmpos.mpos.data.cart.CartDraft
import com.eventmpos.mpos.ui.theme.Brand
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

/** Asks for a label (defaulting to the customer's name) before parking the cart. */
@Composable
fun SaveDraftDialog(defaultLabel: String, onSave: (String) -> Unit, onDismiss: () -> Unit) {
    var label by remember { mutableStateOf(defaultLabel) }
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Save cart as draft") },
        text = { OutlinedTextField(label, { label = it }, label = { Text("Label") }, singleLine = true) },
        confirmButton = { TextButton({ onSave(label) }) { Text("Save") } },
        dismissButton = { TextButton(onDismiss) { Text("Cancel") } },
    )
}

/** Lists the event's drafts. Resuming over a non-empty cart asks first, since it replaces the cart. */
@Composable
fun DraftsDialog(
    drafts: List<CartDraft>,
    cartIsEmpty: Boolean,
    onResume: (String) -> Unit,
    onDelete: (String) -> Unit,
    onDismiss: () -> Unit,
) {
    var confirming by remember { mutableStateOf<CartDraft?>(null) }
    val time = remember { SimpleDateFormat("HH:mm", Locale.getDefault()) }

    confirming?.let { draft ->
        AlertDialog(
            onDismissRequest = { confirming = null },
            title = { Text("Replace current cart?") },
            text = { Text("The items in your cart will be discarded and \"${draft.label}\" loaded instead.") },
            confirmButton = { TextButton({ onResume(draft.id); onDismiss() }) { Text("Replace") } },
            dismissButton = { TextButton({ confirming = null }) { Text("Keep cart") } },
        )
        return
    }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("Drafts") },
        text = {
            if (drafts.isEmpty()) Text("No saved drafts for this event.", color = Brand.TextMuted)
            else Column {
                drafts.forEach { d ->
                    Row(Modifier.fillMaxWidth().padding(vertical = 4.dp), verticalAlignment = Alignment.CenterVertically) {
                        Column(Modifier.weight(1f)) {
                            Text(d.label, fontWeight = FontWeight.SemiBold, color = Brand.Text)
                            Text(
                                "${d.itemCount} items · ${d.customerName ?: "no customer"} · ${time.format(Date(d.createdAtEpochMs))}",
                                fontSize = 12.sp, color = Brand.TextMuted,
                            )
                        }
                        TextButton({ onDelete(d.id) }) { Text("Delete", color = Brand.Danger) }
                        TextButton({ if (cartIsEmpty) { onResume(d.id); onDismiss() } else confirming = d }) { Text("Resume") }
                    }
                }
            }
        },
        confirmButton = { TextButton(onDismiss) { Text("Close") } },
    )
}
