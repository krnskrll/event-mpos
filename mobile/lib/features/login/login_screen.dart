import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/logo.dart';
import '../../app/theme.dart';
import '../../app/widgets.dart';
import '../../core/network/api_failure.dart';
import '../../data/auth/auth_repository.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    // Keyboard autocorrect can add spaces or capitals to the address.
    final email = _email.text.replaceAll(' ', '').toLowerCase();
    if (email.isEmpty || _password.text.isEmpty) {
      setState(() => _error = 'Enter your email and password');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref.read(authRepositoryProvider).login(email, _password.text);
    } on ApiFailure catch (e) {
      if (mounted) {
        setState(() {
          _loading = false;
          _error = e.message;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final form = Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        const Text('Welcome back', style: AppText.title),
        Text('Sign in with your Bhoomy account', style: AppText.body.copyWith(color: Brand.slate)),
        _Field(label: 'Email', controller: _email, keyboard: TextInputType.emailAddress, error: _error != null),
        _Field(label: 'Password', controller: _password, password: true, error: _error != null, onSubmitted: _login),
        if (_error != null)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(_error!, style: AppText.label.copyWith(color: Brand.errorInk)),
          ),
        const SizedBox(height: 24),
        BrandButton(text: 'Sign in', onPressed: _login, loading: _loading),
        const SizedBox(height: 14),
        Text('Trouble signing in? Ask your supervisor.', textAlign: TextAlign.center, style: AppText.caption),
      ],
    );

    if (isWide(context)) {
      return Scaffold(
        body: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(14),
            child: Row(
              children: [
                const SizedBox(width: 540, child: _BrandPanel()),
                Expanded(
                  child: Center(
                    child: SingleChildScrollView(
                      padding: const EdgeInsets.all(24),
                      child: Container(
                        width: 420,
                        padding: const EdgeInsets.all(36),
                        decoration: BoxDecoration(
                          color: Brand.surface,
                          borderRadius: BorderRadius.circular(28),
                          border: Border.all(color: Brand.divider),
                          boxShadow: AppShadows.panel,
                        ),
                        child: form,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }
    // Phone: forest header (light status bar), the form slides up over it as a sheet.
    return AnnotatedRegion(
      value: SystemUiOverlayStyle.light.copyWith(statusBarColor: Colors.transparent),
      child: Scaffold(
        backgroundColor: Brand.forest,
        body: Stack(
          children: [
            const Positioned(
              right: -90,
              top: -60,
              child: CircleAvatar(radius: 150, backgroundColor: Brand.forestRaised),
            ),
            SafeArea(
              bottom: false,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const Padding(padding: EdgeInsets.fromLTRB(28, 32, 28, 28), child: _BrandHeading(compact: true)),
                  Expanded(
                    child: Container(
                      decoration: const BoxDecoration(
                        color: Brand.canvas,
                        borderRadius: BorderRadius.vertical(top: Radius.circular(30)),
                      ),
                      child: SingleChildScrollView(padding: const EdgeInsets.fromLTRB(24, 30, 24, 24), child: form),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Tablet: the forest brand panel beside the sign-in card.
class _BrandPanel extends StatelessWidget {
  const _BrandPanel();

  @override
  Widget build(BuildContext context) {
    const facts = [
      (Icons.cloud_off_outlined, 'Keeps selling when the connection drops'),
      (Icons.print_outlined, 'Order number and receipt the moment you confirm'),
      (Icons.sync, 'Products, prices and promos stay in sync'),
    ];
    return Container(
      clipBehavior: Clip.antiAlias,
      decoration: BoxDecoration(color: Brand.forest, borderRadius: BorderRadius.circular(28)),
      child: Stack(
        children: [
          const Positioned(
            right: -120,
            bottom: -140,
            child: CircleAvatar(radius: 230, backgroundColor: Brand.forestRaised),
          ),
          Padding(
            padding: const EdgeInsets.all(44),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const _BrandHeading(compact: false),
                const Spacer(),
                Text(
                  'Sell at the counter,\nwith or without signal.',
                  style: AppText.title.copyWith(color: Colors.white, fontSize: 36),
                ),
                const SizedBox(height: 22),
                for (final (icon, text) in facts)
                  Padding(
                    padding: const EdgeInsets.only(top: 12),
                    child: Row(
                      children: [
                        Container(
                          width: 34,
                          height: 34,
                          decoration: BoxDecoration(
                            color: Colors.white.withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(11),
                          ),
                          child: Icon(icon, size: 18, color: Brand.gold),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Text(text, style: AppText.body.copyWith(color: Brand.greenPale)),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

/// Logo and name; the phone version adds the tagline under the name.
class _BrandHeading extends StatelessWidget {
  const _BrandHeading({required this.compact});

  final bool compact;

  @override
  Widget build(BuildContext context) {
    final name = Text('Bhoomy POS', style: AppText.title.copyWith(color: Colors.white));
    if (!compact) {
      return Row(
        children: [
          const BhoomyLogo(size: 56, tileColor: Brand.forestRaised),
          const SizedBox(width: 14),
          name,
        ],
      );
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const BhoomyLogo(size: 60, tileColor: Brand.forestRaised),
        const SizedBox(height: 18),
        name,
        Text('Sell at the counter, with or without signal.', style: AppText.body.copyWith(color: Brand.greenTint)),
      ],
    );
  }
}

class _Field extends StatefulWidget {
  const _Field({
    required this.label,
    required this.controller,
    this.keyboard,
    this.password = false,
    this.error = false,
    this.onSubmitted,
  });

  final String label;
  final TextEditingController controller;
  final TextInputType? keyboard;
  final bool password;
  final bool error;
  final VoidCallback? onSubmitted;

  @override
  State<_Field> createState() => _FieldState();
}

class _FieldState extends State<_Field> {
  late bool _hidden = widget.password;

  @override
  Widget build(BuildContext context) {
    final errorLine = OutlineInputBorder(
      borderRadius: BorderRadius.circular(12),
      borderSide: const BorderSide(color: Brand.error),
    );
    return Padding(
      padding: const EdgeInsets.only(top: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(widget.label, style: AppText.label),
          const SizedBox(height: 6),
          TextField(
            controller: widget.controller,
            keyboardType: widget.keyboard,
            obscureText: _hidden,
            autocorrect: false,
            textInputAction: widget.password ? TextInputAction.done : TextInputAction.next,
            onSubmitted: (_) => widget.onSubmitted?.call(),
            decoration: InputDecoration(
              enabledBorder: widget.error ? errorLine : null,
              suffixIcon: widget.password
                  ? IconButton(
                      onPressed: () => setState(() => _hidden = !_hidden),
                      tooltip: _hidden ? 'Show password' : 'Hide password',
                      icon: Icon(
                        _hidden ? Icons.visibility_outlined : Icons.visibility_off_outlined,
                        color: Brand.muted,
                      ),
                    )
                  : null,
            ),
          ),
        ],
      ),
    );
  }
}
