import 'package:flutter/material.dart';
import '../api/auth_api.dart';

/// 登录页面：与员工端同接口（手机号 + 密码），对接后端登录
class LoginPage extends StatefulWidget {
  const LoginPage({super.key});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final TextEditingController _phoneController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void dispose() {
    _phoneController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF5F6FA),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 32),
              const Text(
                '配送员登录',
                style: TextStyle(fontSize: 26, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Text(
                '欢迎使用配送员端，请先完成登录',
                style: TextStyle(fontSize: 14, color: Colors.grey[600]),
              ),
              const SizedBox(height: 40),
              Card(
                elevation: 0,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16,
                    vertical: 24,
                  ),
                  child: Column(
                    children: [
                      TextField(
                        controller: _phoneController,
                        keyboardType: TextInputType.phone,
                        decoration: const InputDecoration(
                          labelText: '手机号',
                          hintText: '请输入11位手机号码',
                          prefixIcon: Icon(Icons.phone_iphone),
                        ),
                      ),
                      const SizedBox(height: 16),
                      TextField(
                        controller: _passwordController,
                        obscureText: true,
                        decoration: const InputDecoration(
                          labelText: '密码',
                          hintText: '请输入登录密码',
                          prefixIcon: Icon(Icons.lock_outline),
                        ),
                      ),
                      const SizedBox(height: 12),
                      if (_errorMessage != null)
                        Align(
                          alignment: Alignment.centerLeft,
                          child: Text(
                            _errorMessage!,
                            style: const TextStyle(
                              color: Colors.redAccent,
                              fontSize: 12,
                            ),
                          ),
                        ),
                      const SizedBox(height: 24),
                      SizedBox(
                        width: double.infinity,
                        height: 48,
                        child: FilledButton(
                          onPressed: _isLoading ? null : _onLoginPressed,
                          child: _isLoading
                              ? const SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    valueColor: AlwaysStoppedAnimation(
                                      Colors.white,
                                    ),
                                  ),
                                )
                              : const Text('登录'),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
              const Spacer(),
              Center(
                child: Text(
                  '登录即表示同意《用户协议》和《隐私政策》',
                  style: TextStyle(fontSize: 12, color: Colors.grey[500]),
                ),
              ),
              const SizedBox(height: 16),
            ],
          ),
        ),
      ),
    );
  }

  void _onLoginPressed() async {
    final phone = _phoneController.text.trim();
    final password = _passwordController.text;

    if (phone.length != 11 || int.tryParse(phone) == null) {
      _showSnackBar('请输入正确的11位手机号码');
      return;
    }

    if (password.isEmpty) {
      _showSnackBar('请输入密码');
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      final response = await AuthApi.login(phone: phone, password: password);

      if (!mounted) return;

      if (response.isSuccess && response.data != null) {
        final employee = response.data!.employee;

        // 只允许配送员登录该 APP
        if (!employee.isDelivery) {
          setState(() {
            _isLoading = false;
            _errorMessage = '当前账号不是配送员，请使用配送员账号登录';
          });
          return;
        }

        // 登录成功后，跳转到主页面（并传递手机号或姓名等）
        Navigator.of(
          context,
        ).pushReplacementNamed('/main', arguments: employee.phone);
      } else {
        setState(() {
          _isLoading = false;
          _errorMessage = response.message.isNotEmpty
              ? response.message
              : '登录失败，请稍后重试';
        });
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        _errorMessage = '登录异常: ${e.toString()}';
      });
    } finally {
      if (mounted && _isLoading) {
        // 正常情况下在成功/失败分支里已经置为 false，这里兜底
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  void _showSnackBar(String message) {
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }
}
