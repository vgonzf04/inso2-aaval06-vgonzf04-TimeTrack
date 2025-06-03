"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";

import { AppSidebar } from "@/components/app-sidebar";
import { DataTable } from "@/components/data-table";
import { SiteHeader } from "@/components/site-header";
import {
  SidebarInset,
  SidebarProvider,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

import CrearEmpleadoForm from "@/components/empleado";
import EliminarEmpleadoForm from "@/components/eliminarEmpleado";

// Función auxiliar para formatear fecha a "YYYY-MM-DD"
function formatYYYYMMDD(d) {
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

export default function AdminPage() {
  const router = useRouter();

  // —— Vacaciones solicitadas de todos los empleados ——
  const [vacaciones, setVacaciones] = useState([]);
  const [loadingVac, setLoadingVac] = useState(true);
  const [errorVac, setErrorVac] = useState(null);

  // —— Fichajes de hoy ——
  const [fichajesAbiertos, setFichajesAbiertos] = useState(0);
  const [fichajesCerrados, setFichajesCerrados] = useState(0);
  const [loadingFich, setLoadingFich] = useState(true);
  const [errorFich, setErrorFich] = useState(null);

  // —— Vacaciones por estado ——
  const [vacPorEstado, setVacPorEstado] = useState([]);
  const [loadingVacEst, setLoadingVacEst] = useState(true);
  const [errorVacEst, setErrorVacEst] = useState(null);

  // —— Horas trabajadas (hoy) ——
  const [horasHoy, setHorasHoy] = useState([]);
  const [loadingHoras, setLoadingHoras] = useState(true);
  const [errorHoras, setErrorHoras] = useState(null);

  // ── useEffect #1: Validar /me y traer Vacaciones de Empleados ──
  useEffect(() => {
    async function fetchVacacionesEmpleados() {
      try {
        // 1) /me -> verifica token
        const meRes = await fetch("http://localhost:3000/me", {
          method: "GET",
          credentials: "include",
        });
        if (!meRes.ok) {
          router.push("/login");
          return;
        }

        // 2) Carga "Vacaciones de empleados"
        const resVac = await fetch("http://localhost:3000/vacaciones/empleados", {
          method: "GET",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
        });
        if (resVac.status === 401 || resVac.status === 403) {
          setLoadingVac(false);
          router.push("/login");
          return;
        }
        if (!resVac.ok) throw new Error("Error al cargar vacaciones de empleados");

        const dataVac = await resVac.json();
        const vacasFlat = dataVac.map((v) => ({
          id: v.id,
          empleado: v.empleado?.nombre || "—",
          fechaInicio: v.inicioStr || "—",
          fechaFin: v.finStr || "—",
          estado: (v.estado || "").toLowerCase(),
        }));
        setVacaciones(vacasFlat);
        setLoadingVac(false);
      } catch (err) {
        console.error(err);
        setErrorVac(err.message || "Algo salió mal cargando vacaciones");
        setLoadingVac(false);
      }
    }

    fetchVacacionesEmpleados();
  }, [router]);

  // ── useEffect #2: Estadísticas (fichajes hoy, vac por estado, horas hoy) ──
  useEffect(() => {
    // A) Fichajes de hoy
    async function fetchFichajesHoy() {
      try {
        const hoy = formatYYYYMMDD(new Date());
        const res = await fetch(
          `http://localhost:3000/dashboard/fichajes-dia?dia=${hoy}`,
          {
            method: "GET",
            credentials: "include",
          }
        );
        if (res.status === 401 || res.status === 403) {
          setLoadingFich(false);
          router.push("/login");
          return;
        }
        if (!res.ok) throw new Error("Error al cargar fichajes de hoy");
        const data = await res.json();
        setFichajesAbiertos(data.fichajes_abiertos ?? 0);
        setFichajesCerrados(data.fichajes_cerrados ?? 0);
        setLoadingFich(false);
      } catch (err) {
        console.error(err);
        setErrorFich(err.message || "Algo falló cargando fichajes del día");
        setLoadingFich(false);
      }
    }

    // B) Vacaciones por estado
    async function fetchVacacionesPorEstado() {
      try {
        const res = await fetch("http://localhost:3000/dashboard/vacaciones-por-estado", {
          method: "GET",
          credentials: "include",
        });
        if (res.status === 401 || res.status === 403) {
          setLoadingVacEst(false);
          router.push("/login");
          return;
        }
        if (!res.ok) throw new Error("Error al cargar vacaciones por estado");
        const data = await res.json();
        // data = [ { estado: "pendiente", cantidad: 3 }, ... ]
        const withId = data.map((x) => ({
          id: x.estado,      // cada estado es único
          estado: x.estado,
          cantidad: x.cantidad,
        }));
        setVacPorEstado(withId);
        setLoadingVacEst(false);
      } catch (err) {
        console.error(err);
        setErrorVacEst(err.message || "Algo falló cargando vacaciones por estado");
        setLoadingVacEst(false);
      }
    }

    // C) Horas trabajadas (hoy)
    async function fetchHorasHoy() {
      try {
        const hoy = formatYYYYMMDD(new Date());
        const res = await fetch(
          `http://localhost:3000/dashboard/horas-periodo?inicio=${hoy}&fin=${hoy}`,
          {
            method: "GET",
            credentials: "include",
          }
        );
        if (res.status === 401 || res.status === 403) {
          setLoadingHoras(false);
          router.push("/login");
          return;
        }
        if (!res.ok) throw new Error("Error al cargar horas trabajadas de hoy");
        const data = await res.json();
        // data = [
        //   { empleado_id: 40, nombre: "...", total_horas: 0.125, total_minutos: 7.50 },
        //   ...
        // ]
        const withId = data.map((x) => ({
          id: x.empleado_id,
          empleado_id: x.empleado_id,
          nombre: x.nombre,
          total_horas: x.total_horas,     // ya viene con 3 decimales desde el backend
          total_minutos: x.total_minutos, // ya viene con 2 decimales desde el backend
        }));
        setHorasHoy(withId);
        setLoadingHoras(false);
      } catch (err) {
        console.error(err);
        setErrorHoras(err.message || "Algo falló cargando horas trabajadas");
        setLoadingHoras(false);
      }
    }

    // Llamamos a las tres funciones en paralelo:
    fetchFichajesHoy();
    fetchVacacionesPorEstado();
    fetchHorasHoy();
  }, [router]);

  // —— Columnas para “Horas Trabajadas (Hoy)” ——
  const horasHoyColumns = [
    { header: "Empleado ID", accessorKey: "empleado_id" },
    { header: "Nombre", accessorKey: "nombre" },
    {
      header: "Total Horas (hoy)",
      accessorKey: "total_horas",
      cell: ({ getValue }) => {
        const h = getValue() ?? 0;
        // Aseguramos 3 decimales en pantalla
        return Number(h).toFixed(3);
      },
    },
    {
      header: "Minutos (hoy)",
      accessorKey: "total_minutos",
      cell: ({ getValue }) => {
        const m = getValue() ?? 0;
        // Aseguramos 2 decimales en pantalla
        return Number(m).toFixed(2);
      },
    },
  ];

  return (
    <SidebarProvider
      style={{
        "--sidebar-width": "calc(var(--spacing) * 72)",
        "--header-height": "calc(var(--spacing) * 12)",
      }}
    >
      <AppSidebar variant="inset" />
      <SidebarInset>
        <SiteHeader />

        <div className="flex flex-1 flex-col gap-4 py-4 px-4 lg:px-6">
          {/* ────── Estadística: Fichajes de Hoy ────── */}
          <section className="bg-white rounded-lg shadow p-4">
            <h3 className="text-xl font-medium mb-2">Fichajes (Hoy)</h3>
            {loadingFich ? (
              <p>Cargando fichajes del día…</p>
            ) : errorFich ? (
              <p className="text-red-600">Error: {errorFich}</p>
            ) : (
              <div className="flex gap-6">
                <div className="flex-1 p-4 bg-gray-100 rounded-lg text-center">
                  <p className="text-sm text-gray-600">Abiertos</p>
                  <p className="text-2xl font-semibold">{fichajesAbiertos}</p>
                </div>
                <div className="flex-1 p-4 bg-gray-100 rounded-lg text-center">
                  <p className="text-sm text-gray-600">Cerrados</p>
                  <p className="text-2xl font-semibold">{fichajesCerrados}</p>
                </div>
              </div>
            )}
          </section>

          {/* ────── Estadística: Vacaciones por Estado ────── */}
          <section className="bg-white rounded-lg shadow p-4">
            <h3 className="text-xl font-medium mb-2">Vacaciones por Estado</h3>
            {loadingVacEst ? (
              <p>Cargando vacaciones por estado…</p>
            ) : errorVacEst ? (
              <p className="text-red-600">Error: {errorVacEst}</p>
            ) : (
              <div className="flex gap-6">
                {vacPorEstado.map((v) => (
                  <div
                    key={v.id}
                    className="flex-1 p-4 bg-gray-100 rounded-lg text-center"
                  >
                    <p className="text-sm text-gray-600 capitalize">{v.estado}</p>
                    <p className="text-2xl font-semibold">{v.cantidad}</p>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* ────── Estadística: Horas Trabajadas (Hoy) ────── */}
          <section className="bg-white rounded-lg shadow p-4">
            <h3 className="text-xl font-medium mb-2">Horas Trabajadas (Hoy)</h3>
            {loadingHoras ? (
              <p>Cargando horas trabajadas…</p>
            ) : errorHoras ? (
              <p className="text-red-600">Error: {errorHoras}</p>
            ) : (
              <div className="overflow-x-auto">
                <DataTable data={horasHoy} columns={horasHoyColumns} />
              </div>
            )}
          </section>

          {/* ────── Tabla Principal: Vacaciones Solicitadas ────── */}
          <section className="mb-8">
            <h2 className="text-2xl font-semibold mb-4">Vacaciones Solicitadas</h2>
            {loadingVac ? (
              <p>Cargando vacaciones de empleados…</p>
            ) : errorVac ? (
              <p className="text-red-600">Error: {errorVac}</p>
            ) : (
              <div className="overflow-x-auto">
                <DataTable
                  data={vacaciones}
                  columns={[
                    { header: "ID", accessorKey: "id" },
                    { header: "Empleado", accessorKey: "empleado" },
                    { header: "Fecha Inicio", accessorKey: "fechaInicio" },
                    { header: "Fecha Fin", accessorKey: "fechaFin" },
                    { header: "Estado", accessorKey: "estado" },
                    {
                      id: "acciones",
                      header: "Acciones",
                      cell: ({ row }) => {
                        const v = row.original;
                        if (v.estado !== "pendiente") return null;
                        return (
                          <div className="flex gap-2">
                            <Button
                              size="sm"
                              variant="outline"
                              onClick={async () => {
                                try {
                                  const res = await fetch(
                                    `http://localhost:3000/vacaciones/${v.id}/aprobar`,
                                    {
                                      method: "PUT",
                                      credentials: "include",
                                    }
                                  );
                                  if (!res.ok) throw new Error("No se pudo aprobar");
                                  const actualizado = await res.json();
                                  setVacaciones((prev) =>
                                    prev.map((x) =>
                                      x.id === actualizado.id
                                        ? {
                                            ...x,
                                            estado: (actualizado.estado || "").toLowerCase(),
                                          }
                                        : x
                                    )
                                  );
                                  toast.success("Vacación aprobada");
                                } catch (e) {
                                  toast.error(e.message || "Error al aprobar");
                                }
                              }}
                            >
                              Aprobar
                            </Button>
                            <Button
                              size="sm"
                              variant="destructive"
                              onClick={async () => {
                                try {
                                  const res = await fetch(
                                    `http://localhost:3000/vacaciones/${v.id}/rechazar`,
                                    {
                                      method: "PUT",
                                      credentials: "include",
                                    }
                                  );
                                  if (!res.ok) throw new Error("No se pudo rechazar");
                                  const actualizado = await res.json();
                                  setVacaciones((prev) =>
                                    prev.map((x) =>
                                      x.id === actualizado.id
                                        ? {
                                            ...x,
                                            estado: (actualizado.estado || "").toLowerCase(),
                                          }
                                        : x
                                    )
                                  );
                                  toast.success("Vacación rechazada");
                                } catch (e) {
                                  toast.error(e.message || "Error al rechazar");
                                }
                              }}
                            >
                              Rechazar
                            </Button>
                          </div>
                        );
                      },
                    },
                  ]}
                />
              </div>
            )}
          </section>

          {/* ────── Formularios para Supervisor ────── */}
          <div className="space-y-6 mt-6 w-full">
            <div className="w-full">
              <CrearEmpleadoForm />
            </div>
            <div className="w-full">
              <EliminarEmpleadoForm />
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
