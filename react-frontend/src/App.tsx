import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './auth/AuthContext';
import { ProtectedRoute } from './auth/ProtectedRoute';
import { Layout } from './components/Layout';
import { LoginPage } from './pages/LoginPage';
import { DashboardPage } from './pages/DashboardPage';
import { TasksPage } from './pages/TasksPage';
import { CampaignsPage } from './pages/CampaignsPage';
import { LeadsPage } from './pages/LeadsPage';
import { AccountsPage } from './pages/AccountsPage';
import { ContactsPage } from './pages/ContactsPage';
import { OpportunitiesPage } from './pages/OpportunitiesPage';
import { SearchPage } from './pages/SearchPage';

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            element={
              <ProtectedRoute>
                <Layout />
              </ProtectedRoute>
            }
          >
            <Route index element={<DashboardPage />} />
            <Route path="tasks" element={<TasksPage />} />
            <Route path="campaigns" element={<CampaignsPage />} />
            <Route path="leads" element={<LeadsPage />} />
            <Route path="accounts" element={<AccountsPage />} />
            <Route path="contacts" element={<ContactsPage />} />
            <Route path="opportunities" element={<OpportunitiesPage />} />
            <Route path="search" element={<SearchPage />} />
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
