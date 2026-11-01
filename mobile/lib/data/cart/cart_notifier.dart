import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../domain/cart.dart';
import '../../domain/customer.dart';
import '../product/product_repository.dart';
import '../promotion/promotion_repository.dart';
import '../session/session.dart';
import 'draft_repository.dart';

final cartProvider = NotifierProvider<CartNotifier, Cart>(CartNotifier.new);

/// The cart priced with the event's products and best promotion.
final pricedCartProvider = Provider.autoDispose(
  (ref) => PricedCart.of(
    ref.watch(cartProvider),
    ref.watch(productsProvider).value ?? const [],
    ref.watch(promotionsProvider).value ?? const [],
    DateTime.now(),
  ),
);

/// Local-only cart for the selected event. Every change is autosaved (debounced); entering an event
/// restores its saved cart, and switching to another event drops the old event's carts.
class CartNotifier extends Notifier<Cart> {
  String? _eventId;
  Timer? _autosave;

  DraftRepository get _drafts => ref.read(draftRepositoryProvider);

  @override
  Cart build() {
    final eventId = ref.watch(sessionEventProvider.select((e) => e?.id));
    if (_eventId != null && eventId != null && eventId != _eventId) _drafts.clearEvent(_eventId!);
    _eventId = eventId;
    if (eventId != null) {
      _drafts.loadLive(eventId).then((saved) {
        if (saved != null && _eventId == eventId && state.isEmpty) state = saved;
      });
    }
    ref.onDispose(() => _autosave?.cancel());
    return const Cart();
  }

  void add(String productId, {String? variantId, int qty = 1}) => adjust(cartKey(productId, variantId), qty);

  void adjust(String key, int delta) {
    final qty = (state.items[key] ?? 0) + delta;
    _update(Cart(items: {...state.items, key: qty}..removeWhere((_, q) => q <= 0), customer: state.customer));
  }

  void setCustomer(Customer? customer) => _update(Cart(items: state.items, customer: customer));

  void clear() => _update(const Cart());

  /// Parks the cart as a named draft and starts a fresh one.
  Future<void> saveDraft(String label) async {
    if (_eventId == null || state.items.isEmpty) return;
    await _drafts.saveDraft(_eventId!, label, state);
    clear();
  }

  Future<void> resumeDraft(String id) async {
    final draft = await _drafts.take(id);
    if (draft != null) _update(draft);
  }

  /// The event closed: its parked carts can no longer be sold.
  Future<void> dropEvent(String eventId) async {
    await _drafts.clearEvent(eventId);
    if (eventId == _eventId) state = const Cart();
  }

  void _update(Cart cart) {
    state = cart;
    final eventId = _eventId;
    if (eventId == null) return;
    _autosave?.cancel();
    _autosave = Timer(const Duration(milliseconds: 300), () => _drafts.saveLive(eventId, cart));
  }
}
