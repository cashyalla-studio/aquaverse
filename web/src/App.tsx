import { Suspense, lazy } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { I18nextProvider } from 'react-i18next'
import i18n from './i18n'
import Layout from './components/layout/Layout'
import { PWAInstallPrompt } from './components/PWAInstallPrompt'
import { useAuthStore } from './store/authStore'
import './index.css'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 5 * 60 * 1000, retry: 2 },
  },
})

// Lazy imports (코드 스플리팅)
const FishEncyclopedia = lazy(() => import('./pages/FishEncyclopedia'))
const FishDetail = lazy(() => import('./pages/FishDetail'))
const Community = lazy(() => import('./pages/Community').then((m) => ({ default: m.default })))
const BoardPage = lazy(() => import('./pages/Community').then((m) => ({ default: m.BoardPage })))
const Marketplace = lazy(() => import('./pages/Marketplace'))
const Login = lazy(() => import('./pages/Login'))
const Register = lazy(() => import('./pages/Register'))
const TradeChat = lazy(() => import('./pages/trade/TradeChat'))
const PhoneVerify = lazy(() => import('./pages/auth/PhoneVerify'))
const TotpSetup = lazy(() => import('./pages/auth/TotpSetup'))
const FishCompatibility = lazy(() => import('./pages/fish/FishCompatibility'))
const TankDoctorPage = lazy(() => import('./pages/tanks/TankDoctorPage'))
const BusinessList = lazy(() => import('./pages/business/BusinessList'))
const BusinessDetail = lazy(() => import('./pages/business/BusinessDetail'))
const MapPage = lazy(() => import('./pages/map/MapPage'))
const VideoFeed = lazy(() => import('./pages/video/VideoFeed'))
const SubscriptionPage = lazy(() => import('./pages/subscription/SubscriptionPage'))
const SocialFeed = lazy(() => import('./pages/social/Feed'))
const AdminDashboard = lazy(() => import('./pages/admin/AdminDashboard'))
const AdminUsers = lazy(() => import('./pages/admin/AdminUsers'))
const PipelinePage = lazy(() => import('./pages/admin/PipelinePage'))
const SpeciesIdentify = lazy(() => import('./pages/species/SpeciesIdentify'))
const BadgesPage = lazy(() => import('./pages/BadgesPage'))
const CareHub = lazy(() => import('./pages/CareHub'))

const Spinner = () => (
  <div className="flex h-64 items-center justify-center">
    <div className="w-8 h-8 border-4 border-primary-200 border-t-primary-500 rounded-full animate-spin" />
  </div>
)

// 인증 필요 라우트 가드
function PrivateRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const location = useLocation()

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />
  }
  return <>{children}</>
}

export default function App() {
  return (
    <I18nextProvider i18n={i18n}>
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>
          <Layout>
            <Suspense fallback={<Spinner />}>
              <Routes>
                <Route path="/" element={<FishEncyclopedia />} />
                <Route path="/fish" element={<FishEncyclopedia />} />
                <Route path="/fish/:id" element={<FishDetail />} />
                <Route path="/fish/:id/compatible" element={<FishCompatibility />} />
                <Route path="/community" element={<Community />} />
                <Route path="/community/:boardID" element={<BoardPage />} />
                <Route path="/marketplace" element={<Marketplace />} />
                <Route path="/marketplace/:id" element={<div className="p-8 text-center text-gray-500">Listing Detail (WIP)</div>} />
                <Route
                  path="/marketplace/create"
                  element={
                    <PrivateRoute>
                      <div className="p-8 text-center text-gray-500">Create Listing (WIP)</div>
                    </PrivateRoute>
                  }
                />
                <Route path="/tanks" element={<div className="p-8 text-center text-gray-500">My Tanks (WIP)</div>} />
                <Route
                  path="/tanks/:id/doctor"
                  element={
                    <PrivateRoute>
                      <TankDoctorPage />
                    </PrivateRoute>
                  }
                />
                <Route path="/businesses" element={<BusinessList />} />
                <Route path="/businesses/:id" element={<BusinessDetail />} />
                <Route path="/map" element={<MapPage />} />
                <Route path="/videos" element={<VideoFeed />} />
                <Route path="/subscription" element={<SubscriptionPage />} />
                <Route
                  path="/feed"
                  element={
                    <PrivateRoute>
                      <SocialFeed />
                    </PrivateRoute>
                  }
                />
                <Route
                  path="/admin"
                  element={
                    <PrivateRoute>
                      <AdminDashboard />
                    </PrivateRoute>
                  }
                />
                <Route
                  path="/admin/users"
                  element={
                    <PrivateRoute>
                      <AdminUsers />
                    </PrivateRoute>
                  }
                />
                <Route
                  path="/admin/pipeline"
                  element={
                    <PrivateRoute>
                      <PipelinePage />
                    </PrivateRoute>
                  }
                />
                <Route path="/identify" element={<SpeciesIdentify />} />
                <Route path="/badges" element={<BadgesPage />} />
                <Route
                  path="/care"
                  element={
                    <PrivateRoute>
                      <CareHub />
                    </PrivateRoute>
                  }
                />
                <Route path="/login" element={<Login />} />
                <Route path="/register" element={<Register />} />
                <Route
                  path="/trades/:tradeId/chat"
                  element={
                    <PrivateRoute>
                      <TradeChat />
                    </PrivateRoute>
                  }
                />
                <Route
                  path="/phone/verify"
                  element={
                    <PrivateRoute>
                      <PhoneVerify />
                    </PrivateRoute>
                  }
                />
                <Route
                  path="/settings/security"
                  element={
                    <PrivateRoute>
                      <TotpSetup />
                    </PrivateRoute>
                  }
                />
              </Routes>
            </Suspense>
          </Layout>
          <PWAInstallPrompt />
        </BrowserRouter>
      </QueryClientProvider>
    </I18nextProvider>
  )
}
