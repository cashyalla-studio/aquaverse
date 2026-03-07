import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'features/fish/presentation/fish_list_screen.dart';
import 'features/fish/presentation/fish_detail_screen.dart';
import 'features/marketplace/presentation/marketplace_screen.dart';
import 'features/marketplace/presentation/listing_detail_screen.dart';
import 'features/marketplace/presentation/create_listing_screen.dart';
import 'features/community/presentation/community_screen.dart';
import 'features/tanks/presentation/tanks_screen.dart';
import 'features/auth/presentation/auth_provider.dart';
import 'features/auth/presentation/login_screen.dart';
import 'features/auth/presentation/register_screen.dart';
import 'l10n/app_localizations.dart';

const _supportedLocales = [
  Locale('ko'),
  Locale('en', 'US'),
  Locale('en', 'GB'),
  Locale('en', 'AU'),
  Locale('ja'),
  Locale.fromSubtags(languageCode: 'zh', scriptCode: 'Hans'),
  Locale.fromSubtags(languageCode: 'zh', scriptCode: 'Hant'),
  Locale('de'),
  Locale('fr', 'FR'),
  Locale('fr', 'CA'),
  Locale('es'),
  Locale('pt'),
  Locale('ar'),
  Locale('he'),
];

bool _requiresAuth(String location) {
  if (location == '/marketplace/create') return true;
  // /community/:boardId/create
  if (location.contains('/community/') && location.endsWith('/create')) return true;
  return false;
}

GoRouter _buildRouter(AuthChangeNotifier authNotifier, ProviderContainer container) {
  return GoRouter(
    initialLocation: '/',
    refreshListenable: authNotifier,
    redirect: (context, state) {
      final authState = container.read(authProvider);

      // While auth is still initialising, do not redirect
      if (authState.isLoading) return null;

      final isAuthenticated = authState.isAuthenticated;
      final location = state.uri.toString();
      final isOnLogin = location.startsWith('/login');
      final isOnRegister = location.startsWith('/register');

      // Redirect authenticated users away from /login and /register
      if (isAuthenticated && (isOnLogin || isOnRegister)) {
        return '/';
      }

      // Redirect unauthenticated users away from protected routes
      if (!isAuthenticated && _requiresAuth(location)) {
        final encoded = Uri.encodeComponent(location);
        return '/login?from=$encoded';
      }

      return null;
    },
    routes: [
      // Auth routes (outside ShellRoute so they have no bottom nav)
      GoRoute(path: '/login', builder: (_, __) => const LoginScreen()),
      GoRoute(path: '/register', builder: (_, __) => const RegisterScreen()),

      ShellRoute(
        builder: (context, state, child) => MainScaffold(child: child),
        routes: [
          GoRoute(path: '/', builder: (_, __) => const FishListScreen()),
          GoRoute(path: '/fish', builder: (_, __) => const FishListScreen()),
          GoRoute(
            path: '/fish/:id',
            builder: (_, state) => FishDetailScreen(
              fishId: int.tryParse(state.pathParameters['id']!) ?? 0,
            ),
          ),
          GoRoute(path: '/community', builder: (_, __) => const CommunityScreen()),
          GoRoute(
            path: '/community/:boardId',
            builder: (_, state) => BoardScreen(
              boardId: int.tryParse(state.pathParameters['boardId']!) ?? 0,
              board: state.extra as dynamic,
            ),
          ),
          GoRoute(
            path: '/community/:boardId/posts/:postId',
            builder: (_, state) => PostDetailScreen(
              boardId: int.tryParse(state.pathParameters['boardId']!) ?? 0,
              postId: int.tryParse(state.pathParameters['postId']!) ?? 0,
            ),
          ),
          GoRoute(
            path: '/community/:boardId/create',
            builder: (_, state) => CreatePostScreen(
              boardId: int.tryParse(state.pathParameters['boardId']!) ?? 0,
            ),
          ),
          GoRoute(path: '/marketplace', builder: (_, __) => const MarketplaceScreen()),
          GoRoute(path: '/marketplace/create', builder: (_, __) => const CreateListingScreen()),
          GoRoute(
            path: '/marketplace/:id',
            builder: (_, state) => ListingDetailScreen(
              listingId: int.tryParse(state.pathParameters['id']!) ?? 0,
            ),
          ),
          GoRoute(path: '/tanks', builder: (_, __) => const TanksScreen()),
        ],
      ),
    ],
  );
}

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final prefs = await SharedPreferences.getInstance();
  final savedLocale = prefs.getString('av_locale');

  // Create a ProviderContainer so we can pass it to the router before runApp
  final container = ProviderContainer();

  // Read the authChangeNotifier from the container
  final authNotifier = container.read(authChangeNotifierProvider);

  final router = _buildRouter(authNotifier, container);

  runApp(
    UncontrolledProviderScope(
      container: container,
      child: AquaVerseApp(initialLocale: savedLocale, router: router),
    ),
  );
}

class AquaVerseApp extends ConsumerStatefulWidget {
  final String? initialLocale;
  final GoRouter router;

  const AquaVerseApp({super.key, this.initialLocale, required this.router});

  @override
  ConsumerState<AquaVerseApp> createState() => _AquaVerseAppState();
}

class _AquaVerseAppState extends ConsumerState<AquaVerseApp> {
  late Locale _locale;

  @override
  void initState() {
    super.initState();
    _locale = _parseLocale(widget.initialLocale) ?? const Locale('en', 'US');
  }

  Locale _parseLocale(String? tag) {
    if (tag == null) return const Locale('en', 'US');
    final parts = tag.split('-');
    if (parts.length >= 2) return Locale(parts[0], parts[1]);
    return Locale(parts[0]);
  }

  bool get _isRTL => _locale.languageCode == 'ar' || _locale.languageCode == 'he';

  @override
  Widget build(BuildContext context) => MaterialApp.router(
    title: 'AquaVerse',
    debugShowCheckedModeBanner: false,
    locale: _locale,
    supportedLocales: _supportedLocales,
    localizationsDelegates: const [
      AppLocalizations.delegate,
      GlobalMaterialLocalizations.delegate,
      GlobalWidgetsLocalizations.delegate,
      GlobalCupertinoLocalizations.delegate,
    ],
    routerConfig: widget.router,
    theme: ThemeData(
      colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF0EA5E9)),
      useMaterial3: true,
      cardTheme: CardTheme(
        elevation: 2,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: Color(0xFF0EA5E9),
        foregroundColor: Colors.white,
        elevation: 0,
      ),
    ),
    builder: (context, child) => Directionality(
      textDirection: _isRTL ? TextDirection.rtl : TextDirection.ltr,
      child: child!,
    ),
  );
}

class MainScaffold extends StatefulWidget {
  final Widget child;
  const MainScaffold({super.key, required this.child});

  @override
  State<MainScaffold> createState() => _MainScaffoldState();
}

class _MainScaffoldState extends State<MainScaffold> {
  int _tabIndex = 0;

  static const _tabs = [
    (icon: Icons.home_outlined, activeIcon: Icons.home, path: '/'),
    (icon: Icons.set_meal_outlined, activeIcon: Icons.set_meal, path: '/fish'),
    (icon: Icons.forum_outlined, activeIcon: Icons.forum, path: '/community'),
    (icon: Icons.store_outlined, activeIcon: Icons.store, path: '/marketplace'),
    (icon: Icons.water_outlined, activeIcon: Icons.water, path: '/tanks'),
  ];

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final labels = [l10n.navHome, l10n.navEncyclopedia, l10n.navCommunity, l10n.navMarketplace, l10n.navMyTanks];

    return Scaffold(
      body: widget.child,
      bottomNavigationBar: NavigationBar(
        selectedIndex: _tabIndex,
        onDestinationSelected: (i) {
          setState(() => _tabIndex = i);
          context.go(_tabs[i].path);
        },
        destinations: List.generate(_tabs.length, (i) => NavigationDestination(
          icon: Icon(_tabs[i].icon),
          selectedIcon: Icon(_tabs[i].activeIcon, color: const Color(0xFF0EA5E9)),
          label: labels[i],
        )),
        backgroundColor: Colors.white,
        indicatorColor: const Color(0xFFE0F2FE),
      ),
    );
  }
}
