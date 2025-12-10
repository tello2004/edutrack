import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import Login from "../pages/Login";
import Dashboard from "../pages/Dashboard";
import Alumnos from "../pages/Alumnos";
import Grupos from "../pages/Grupos";
import Asistencias from "../pages/Asistencias";
import Evaluaciones from "../pages/Evaluaciones";
import Calificaciones from "../pages/Calificaciones";
import Reportes from "../pages/Reportes";
import Perfil from "../pages/Perfil";
import Layout from "../components/ui/Layout";


export default function AppRouter() {
return (
<BrowserRouter>
<Routes>
<Route path="/login" element={<Login />} />


{}
<Route element={<Layout />}>
<Route path="/" element={<Dashboard />} />
<Route path="/alumnos" element={<Alumnos />} />
<Route path="/grupos" element={<Grupos />} />
<Route path="/asistencias" element={<Asistencias />} />
<Route path="/evaluaciones" element={<Evaluaciones />} />
<Route path="/calificaciones" element={<Calificaciones />} />
<Route path="/reportes" element={<Reportes />} />
<Route path="/perfil" element={<Perfil />} />
</Route>


<Route path="*" element={<Navigate to="/" />} />
</Routes>
</BrowserRouter>
);
}