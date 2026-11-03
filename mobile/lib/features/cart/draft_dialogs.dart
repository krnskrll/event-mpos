import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../data/cart/cart_notifier.dart';
import '../../data/cart/draft_repository.dart';
import '../../data/session/session.dart';

final draftsProvider = StreamProvider.autoDispose(
  (ref) => ref.watch(draftRepositoryProvider).watchDrafts(ref.watch(sessionEventProvider)!.id),
);

/// Drafts: hold the current order under a name, or resume one. A sheet on phones, a dialog on tablets.
Future<void> showDraftsDialog(BuildContext context, WidgetRef ref) => isWide(context)
    ? showDialog(
        context: context,
        builder: (_) => Dialog(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 480),
            child: const Padding(padding: EdgeInsets.only(top: 12), child: _DraftsSheet()),
          ),
        ),
      )
    : showModalBottomSheet(
        context: context,
        useRootNavigator: true,
        isScrollControlled: true,
        useSafeArea: true,
        builder: (_) => const _DraftsSheet(),
      );

class _DraftsSheet extends ConsumerStatefulWidget {
  const _DraftsSheet();

  @override
  ConsumerState<_DraftsSheet> createState() => _DraftsSheetState();
}

class _DraftsSheetState extends ConsumerState<_DraftsSheet> {
  late final _label = TextEditingController(text: ref.read(cartProvider).customer?.name);

  @override
  void dispose() {
    _label.dispose();
    super.dispose();
  }

  Future<void> _hold() async {
    final messenger = ScaffoldMessenger.of(context);
    await ref.read(cartProvider.notifier).saveDraft(_label.text);
    if (mounted) Navigator.pop(context);
    messenger.showSnackBar(const SnackBar(content: Text('Order put on hold')));
  }

  Future<void> _resume(CartDraft draft) async {
    final replace = ref.read(cartProvider).items.isEmpty || await _confirmReplace(context, draft.label);
    if (!replace) return;
    await ref.read(cartProvider.notifier).resumeDraft(draft.id);
    if (mounted) Navigator.pop(context);
  }

  @override
  Widget build(BuildContext context) {
    final drafts = ref.watch(draftsProvider).value ?? const [];
    final count = ref.watch(cartProvider.select((c) => c.count));
    return Padding(
      padding: EdgeInsets.fromLTRB(20, 8, 20, 20 + MediaQuery.viewInsetsOf(context).bottom),
      child: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              crossAxisAlignment: CrossAxisAlignment.baseline,
              textBaseline: TextBaseline.alphabetic,
              children: [
                const Expanded(child: Text('Drafts', style: AppText.title)),
                Text('Orders put on hold', style: AppText.caption),
              ],
            ),
            if (count > 0) ...[
              const SizedBox(height: 14),
              Container(
                padding: const EdgeInsets.all(14),
                decoration: BoxDecoration(color: Brand.greenSoft, borderRadius: BorderRadius.circular(18)),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      'Hold the current order · ${itemCount(count)}',
                      style: AppText.label.copyWith(color: Brand.forest),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        Expanded(
                          child: TextField(
                            controller: _label,
                            textCapitalization: TextCapitalization.sentences,
                            onSubmitted: (_) => _hold(),
                            decoration: const InputDecoration(hintText: 'Name this draft', isDense: true),
                          ),
                        ),
                        const SizedBox(width: 8),
                        FilledButton(
                          onPressed: _hold,
                          style: FilledButton.styleFrom(minimumSize: const Size(0, 48)),
                          child: const Text('Save'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
            const SizedBox(height: 18),
            Text(
              'SAVED · ${drafts.length}',
              style: AppText.caption.copyWith(color: Brand.slate, fontWeight: FontWeight.w700, letterSpacing: 0.5),
            ),
            const SizedBox(height: 8),
            if (drafts.isEmpty)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 16),
                child: Text('No drafts for this event yet.', style: AppText.body.copyWith(color: Brand.muted)),
              ),
            for (final d in drafts)
              Container(
                margin: const EdgeInsets.only(bottom: 10),
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Brand.surface,
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(color: Brand.cardLine),
                  boxShadow: AppShadows.card,
                ),
                child: Row(
                  children: [
                    CircleAvatar(
                      radius: 22,
                      backgroundColor: d.cart.customer == null ? Brand.sunken : Brand.green,
                      child: Text(
                        d.cart.customer == null ? '?' : initials(d.cart.customer!.name),
                        style: AppText.label.copyWith(
                          color: d.cart.customer == null ? Brand.slate : Colors.white,
                          fontWeight: FontWeight.w900,
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            d.label.isEmpty ? 'Untitled' : d.label,
                            style: AppText.body.copyWith(fontWeight: FontWeight.w700, color: Brand.forest),
                          ),
                          Text(
                            '${itemCount(d.cart.count)} · ${d.cart.customer?.name ?? 'no customer'} · '
                            '${TimeOfDay.fromDateTime(d.createdAt).format(context)}',
                            maxLines: 2,
                            overflow: TextOverflow.ellipsis,
                            style: AppText.caption,
                          ),
                        ],
                      ),
                    ),
                    IconButton(
                      onPressed: () => ref.read(draftRepositoryProvider).delete(d.id),
                      tooltip: 'Delete draft',
                      icon: const Icon(Icons.delete_outline, color: Brand.errorInk),
                    ),
                    FilledButton(
                      onPressed: () => _resume(d),
                      style: FilledButton.styleFrom(minimumSize: const Size(0, 40)),
                      child: const Text('Resume'),
                    ),
                  ],
                ),
              ),
            const SizedBox(height: 4),
            Text(
              'Resuming replaces the current order, so hold it first if you need it. '
              'Drafts stay on this device until the event closes.',
              style: AppText.caption,
            ),
          ],
        ),
      ),
    );
  }
}

Future<bool> _confirmReplace(BuildContext context, String label) async =>
    await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Replace the current order?'),
        content: Text('Its items will be discarded and "$label" loaded instead.'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('Keep order')),
          TextButton(onPressed: () => Navigator.pop(context, true), child: const Text('Replace')),
        ],
      ),
    ) ??
    false;
