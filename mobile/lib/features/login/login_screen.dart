import 'package:flutter/material.dart';
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
    return Scaffold(
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 32),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 420),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Row(
                  children: [
                    BhoomyLogo(size: 48),
                    SizedBox(width: 12),
                    Text('Bhoomy POS', style: AppText.subhead),
                  ],
                ),
                const SizedBox(height: 40),
                const Text('Welcome back', style: AppText.title),
                Text('Sign in with your Bhoomy account', style: AppText.body.copyWith(color: Brand.slate)),
                const SizedBox(height: 16),
                _Field(label: 'Email', controller: _email, keyboard: TextInputType.emailAddress),
                _Field(label: 'Password', controller: _password, password: true, onSubmitted: _login),
                if (_error != null)
                  Padding(
                    padding: const EdgeInsets.only(top: 8),
                    child: Text(_error!, style: AppText.label.copyWith(color: Brand.errorInk)),
                  ),
                const SizedBox(height: 24),
                BrandButton(text: 'Sign in', onPressed: _login, loading: _loading),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _Field extends StatelessWidget {
  const _Field({required this.label, required this.controller, this.keyboard, this.password = false, this.onSubmitted});

  final String label;
  final TextEditingController controller;
  final TextInputType? keyboard;
  final bool password;
  final VoidCallback? onSubmitted;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(label, style: AppText.label),
          const SizedBox(height: 6),
          TextField(
            controller: controller,
            keyboardType: keyboard,
            obscureText: password,
            autocorrect: false,
            textInputAction: password ? TextInputAction.done : TextInputAction.next,
            onSubmitted: (_) => onSubmitted?.call(),
          ),
        ],
      ),
    );
  }
}
