import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../app/theme.dart';
import '../../app/widgets.dart';
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
    } on LoginFailure catch (e) {
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
      backgroundColor: Colors.white,
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.symmetric(horizontal: 24),
          child: Column(
            children: [
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(color: Brand.navy, borderRadius: BorderRadius.circular(20)),
                child: const Icon(Icons.spa_outlined, color: Brand.teal, size: 40),
              ),
              const SizedBox(height: 14),
              const Text(
                'BEAUTY POS',
                style: TextStyle(color: Brand.teal, fontSize: 12, fontWeight: FontWeight.bold, letterSpacing: 3),
              ),
              const SizedBox(height: 24),
              const Text(
                'Welcome Back',
                style: TextStyle(color: Brand.heading, fontSize: 26, fontWeight: FontWeight.bold),
              ),
              const Text('Sign in to continue', style: TextStyle(color: Brand.textFaint, fontSize: 15)),
              const SizedBox(height: 24),
              _Field(label: 'Email', controller: _email, keyboard: TextInputType.emailAddress),
              _Field(label: 'Password', controller: _password, password: true, onSubmitted: _login),
              if (_error != null)
                Padding(
                  padding: const EdgeInsets.only(top: 8),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text(_error!, style: const TextStyle(color: Brand.danger, fontSize: 13)),
                  ),
                ),
              const SizedBox(height: 20),
              BrandButton(text: 'Sign In', onPressed: _login, loading: _loading),
              const SizedBox(height: 28),
              const Text('Powered by event-mpos', style: TextStyle(color: Brand.textGhost, fontSize: 12)),
            ],
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
    final border = OutlineInputBorder(borderRadius: BorderRadius.circular(12));
    return Padding(
      padding: const EdgeInsets.only(top: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: const TextStyle(color: Color(0xFF374151), fontSize: 14, fontWeight: FontWeight.w500),
          ),
          const SizedBox(height: 6),
          TextField(
            controller: controller,
            keyboardType: keyboard,
            obscureText: password,
            autocorrect: false,
            textInputAction: password ? TextInputAction.done : TextInputAction.next,
            onSubmitted: (_) => onSubmitted?.call(),
            decoration: InputDecoration(
              filled: true,
              fillColor: Brand.fieldBg,
              enabledBorder: border.copyWith(borderSide: const BorderSide(color: Brand.fieldBorder)),
              focusedBorder: border.copyWith(borderSide: const BorderSide(color: Brand.navy)),
            ),
          ),
        ],
      ),
    );
  }
}
