import 'package:go_router/go_router.dart';
import 'package:flutter/material.dart';

final appRouter = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(
      path: '/',
      builder: (context, state) => const Scaffold(body: Center(child: Text('AquaVerse Home (WIP)'))),
    ),
    GoRoute(
      path: '/fish',
      builder: (context, state) => const Scaffold(body: Center(child: Text('Fish Encyclopedia (WIP)'))),
      routes: [
        GoRoute(
          path: ':id',
          builder: (context, state) => Scaffold(
            body: Center(child: Text('Fish Detail: ${state.pathParameters['id']} (WIP)')),
          ),
        ),
      ],
    ),
    GoRoute(
      path: '/community',
      builder: (context, state) => const Scaffold(body: Center(child: Text('Community (WIP)'))),
      routes: [
        GoRoute(
          path: ':boardId',
          builder: (context, state) => Scaffold(
            body: Center(child: Text('Board: ${state.pathParameters['boardId']} (WIP)')),
          ),
          routes: [
            GoRoute(
              path: 'post/:postId',
              builder: (context, state) => Scaffold(
                body: Center(child: Text('Post: ${state.pathParameters['postId']} (WIP)')),
              ),
            ),
          ],
        ),
      ],
    ),
    GoRoute(
      path: '/marketplace',
      builder: (context, state) => const Scaffold(body: Center(child: Text('Marketplace (WIP)'))),
      routes: [
        GoRoute(
          path: ':listingId',
          builder: (context, state) => Scaffold(
            body: Center(child: Text('Listing: ${state.pathParameters['listingId']} (WIP)')),
          ),
        ),
        GoRoute(
          path: 'create',
          builder: (context, state) => const Scaffold(body: Center(child: Text('Create Listing (WIP)'))),
        ),
      ],
    ),
    GoRoute(
      path: '/tanks',
      builder: (context, state) => const Scaffold(body: Center(child: Text('My Tanks (WIP)'))),
    ),
    GoRoute(
      path: '/login',
      builder: (context, state) => const Scaffold(body: Center(child: Text('Login (WIP)'))),
    ),
    GoRoute(
      path: '/profile/:userId',
      builder: (context, state) => Scaffold(
        body: Center(child: Text('Profile: ${state.pathParameters['userId']} (WIP)')),
      ),
    ),
  ],
);
