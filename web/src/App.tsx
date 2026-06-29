import { Navigate, Route, Routes } from "react-router-dom";
import { useAuth } from "./auth/AuthContext";
import { ProtectedRoute, homePathForRole } from "./auth/ProtectedRoute";
import LoginPage from "./pages/LoginPage";
import AdminDashboard from "./pages/AdminDashboard";
import ProductsPage from "./pages/ProductsPage";
import WarehousePage from "./pages/WarehousePage";
import BaPage from "./pages/BaPage";

export default function App() {
  const { user } = useAuth();

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />

      <Route
        path="/admin"
        element={
          <ProtectedRoute allow={["ADMIN"]}>
            <AdminDashboard />
          </ProtectedRoute>
        }
      />
      <Route
        path="/admin/products"
        element={
          <ProtectedRoute allow={["ADMIN"]}>
            <ProductsPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/warehouse"
        element={
          <ProtectedRoute allow={["WAREHOUSE", "ADMIN"]}>
            <WarehousePage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/ba"
        element={
          <ProtectedRoute allow={["BA"]}>
            <BaPage />
          </ProtectedRoute>
        }
      />

      {/* Root sends each user to their role's home (or login). */}
      <Route
        path="*"
        element={<Navigate to={user ? homePathForRole(user.role) : "/login"} replace />}
      />
    </Routes>
  );
}
