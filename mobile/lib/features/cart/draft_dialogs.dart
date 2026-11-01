import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../data/cart/cart_notifier.dart';
import '../../data/cart/draft_repository.dart';
import '../../data/session/session.dart';

final _draftsProvider = StreamProvider.autoDispose(
  (ref) => ref.watch(draftRepositoryProvider).watchDrafts(ref.watch(sessionEventProvider)!.id),
);

/// Asks for a label (the customer's name by default) before parking the cart.
Future<void> showSaveDraftDialog(BuildContext context, WidgetRef ref) async {
  final label = TextEditingController(text: ref.read(cartProvider).customer?.name);
  final save = await showDialog<bool>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('Save cart as draft'),
      content: TextField(
        controller: label,
        autofocus: true,
        decoration: const InputDecoration(labelText: 'Label'),
      ),
      actions: [
        TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancel')),
        TextButton(onPressed: () => Navigator.pop(context, true), child: const Text('Save')),
      ],
    ),
  );
  if (save == true) await ref.read(cartProvider.notifier).saveDraft(label.text);
}

/// The event's drafts. Resuming over a non-empty cart asks first, since it replaces the cart.
Future<void> showDraftsDialog(BuildContext context, WidgetRef ref) => showDialog(
  context: context,
  builder: (context) => Consumer(
    builder: (context, ref, _) {
      final drafts = ref.watch(_draftsProvider).value ?? const [];

      Future<void> resume(CartDraft draft) async {
        final replace = ref.read(cartProvider).items.isEmpty || await _confirmReplace(context, draft.label);
        if (!replace) return;
        await ref.read(cartProvider.notifier).resumeDraft(draft.id);
        if (context.mounted) Navigator.pop(context);
      }

      return AlertDialog(
        title: const Text('Drafts'),
        content: drafts.isEmpty
            ? const Text('No saved drafts for this event.', style: TextStyle(color: Brand.textMuted))
            : Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  for (final d in drafts)
                    ListTile(
                      contentPadding: EdgeInsets.zero,
                      title: Text(d.label, style: const TextStyle(fontWeight: FontWeight.w600)),
                      subtitle: Text(
                        '${d.cart.count} items · ${d.cart.customer?.name ?? 'no customer'} · '
                        '${d.createdAt.hour.toString().padLeft(2, '0')}:${d.createdAt.minute.toString().padLeft(2, '0')}',
                      ),
                      onTap: () => resume(d),
                      trailing: IconButton(
                        onPressed: () => ref.read(draftRepositoryProvider).delete(d.id),
                        icon: const Icon(Icons.delete_outline, color: Brand.danger),
                      ),
                    ),
                ],
              ),
        actions: [TextButton(onPressed: () => Navigator.pop(context), child: const Text('Close'))],
      );
    },
  ),
);

Future<bool> _confirmReplace(BuildContext context, String label) async =>
    await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Replace current cart?'),
        content: Text('The items in your cart will be discarded and "$label" loaded instead.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Keep cart')),
          TextButton(onPressed: () => Navigator.pop(context, true), child: const Text('Replace')),
        ],
      ),
    ) ??
    false;
