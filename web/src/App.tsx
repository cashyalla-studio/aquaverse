import { Suspense, lazy } from 'react'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { I18nextProvider } from 'react-i18next'
import i18n from './i18n'
import Layout from './components/layout/Layout'
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

const Spinner = () => (
  <div className="flex h-64 items-center justify-center">
    <div className="w-8 h-8 border-4 border-primary-200 border-t-primary-500 rounded-full animate-spin" />
  </div>
)

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
                <Route path="/community" element={<Community />} />
                <Route path="/community/:boardID" element={<BoardPage />} />
                <Route path="/marketplace" element={<Marketplace />} />
                <Route path="/marketplace/:id" element={<div className="p-8 text-center text-gray-500">Listing Detail (WIP)</div>} />
                <Route path="/marketplace/create" element={<div className="p-8 text-center text-gray-500">Create Listing (WIP)</div>} />
                <Route path="/tanks" element={<div className="p-8 text-center text-gray-500">My Tanks (WIP)</div>} />
                <Route path="/login" element={<div className="p-8 text-center text-gray-500">Login (WIP)</div>} />
              </Routes>
            </Suspense>
          </Layout>
        </BrowserRouter>
      </QueryClientProvider>
    </I18nextProvider>
  )
}
