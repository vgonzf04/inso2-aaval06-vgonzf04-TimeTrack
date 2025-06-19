// frontend/app/admin/page.js

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

  // —— Tarifa €/h ——  
  const [tarifa, setTarifa] = useState(15);

  // ── useEffect #1: validar /me y cargar vacaciones ──
  useEffect(() => {
    async function fetchVacaciones() {
      try {
        const me = await fetch("http://localhost:3000/me", {
          credentials: "include",
        });
        if (!me.ok) return router.push("/login");

        const res = await fetch("http://localhost:3000/vacaciones/empleados", {
          credentials: "include",
        });
        if (res.status === 401 || res.status === 403) {
          setLoadingVac(false);
          return router.push("/login");
        }
        if (!res.ok) throw new Error();

        const data = await res.json();
        setVacaciones(
          data.map((v) => ({
            id: v.id,
            empleado: v.empleado?.nombre || "—",
            fechaInicio: v.inicioStr || "—",
            fechaFin: v.finStr || "—",
            estado: (v.estado || "").toLowerCase(),
          }))
        );
        setLoadingVac(false);
      } catch (e) {
        console.error(e);
        setErrorVac("Error cargando vacaciones");
        setLoadingVac(false);
      }
    }
    fetchVacaciones();
  }, [router]);

  // ── useEffect #2: estadísticas ──
  useEffect(() => {
    // fichajes hoy
    (async () => {
      try {
        const dia = formatYYYYMMDD(new Date());
        const res = await fetch(
          `http://localhost:3000/dashboard/fichajes-dia?dia=${dia}`,
          { credentials: "include" }
        );
        if (!res.ok) {
          setLoadingFich(false);
          return router.push("/login");
        }
        const { fichajes_abiertos, fichajes_cerrados } = await res.json();
        setFichajesAbiertos(fichajes_abiertos || 0);
        setFichajesCerrados(fichajes_cerrados || 0);
      } catch {
        setErrorFich("Error cargando fichajes");
      } finally {
        setLoadingFich(false);
      }
    })();

    // vacaciones por estado
    (async () => {
      try {
        const res = await fetch(
          "http://localhost:3000/dashboard/vacaciones-por-estado",
          { credentials: "include" }
        );
        if (!res.ok) {
          setLoadingVacEst(false);
          return router.push("/login");
        }
        const data = await res.json();
        setVacPorEstado(
          data.map((x) => ({
            id: x.estado,
            estado: x.estado,
            cantidad: x.cantidad,
          }))
        );
      } catch {
        setErrorVacEst("Error cargando estado");
      } finally {
        setLoadingVacEst(false);
      }
    })();

    // horas trabajadas
    (async () => {
      try {
        const dia = formatYYYYMMDD(new Date());
        const res = await fetch(
          `http://localhost:3000/dashboard/horas-periodo?inicio=${dia}&fin=${dia}`,
          { credentials: "include" }
        );
        if (!res.ok) {
          setLoadingHoras(false);
          return router.push("/login");
        }
        const data = await res.json();
        setHorasHoy(
          data.map((x) => ({
            id: x.empleado_id,
            empleado_id: x.empleado_id,
            nombre: x.nombre,
            total_horas: x.total_horas,
            total_minutos: x.total_minutos,
          }))
        );
      } catch {
        setErrorHoras("Error cargando horas");
      } finally {
        setLoadingHoras(false);
      }
    })();
  }, [router]);

  const horasHoyColumns = [
    { header: "Empleado ID", accessorKey: "empleado_id" },
    { header: "Nombre", accessorKey: "nombre" },
    {
      header: "Total Horas",
      accessorKey: "total_horas",
      cell: ({ getValue }) => Number(getValue() ?? 0).toFixed(3),
    },
    {
      header: "Minutos",
      accessorKey: "total_minutos",
      cell: ({ getValue }) => Number(getValue() ?? 0).toFixed(2),
    },
    {
      header: "€/h",
      accessorKey: "tarifa",
      cell: () => tarifa.toFixed(2),
    },
    {
      header: "Total €",
      accessorKey: "total_horas",
      cell: ({ getValue }) => (getValue() * tarifa).toFixed(2),
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
          {/* Fichajes de Hoy */}
          <section className="bg-white rounded-lg shadow p-4">
            <h3 className="text-xl font-medium mb-2">Fichajes (Hoy)</h3>
            {loadingFich ? (
              <p>Cargando fichajes…</p>
            ) : errorFich ? (
              <p className="text-red-600">{errorFich}</p>
            ) : (
              <div className="flex gap-6">
                <div className="flex-1 p-4 bg-gray-100 rounded text-center">
                  <p className="text-sm text-gray-600">Abiertos</p>
                  <p className="text-2xl font-semibold">
                    {fichajesAbiertos}
                  </p>
                </div>
                <div className="flex-1 p-4 bg-gray-100 rounded text-center">
                  <p className="text-sm text-gray-600">Cerrados</p>
                  <p className="text-2xl font-semibold">
                    {fichajesCerrados}
                  </p>
                </div>
              </div>
            )}
          </section>

          {/* Vacaciones por Estado */}
          <section className="bg-white rounded-lg shadow p-4">
            <h3 className="text-xl font-medium mb-2">
              Vacaciones por Estado
            </h3>
            {loadingVacEst ? (
              <p>Cargando…</p>
            ) : errorVacEst ? (
              <p className="text-red-600">{errorVacEst}</p>
            ) : (
              <div className="flex gap-6">
                {vacPorEstado.map((v) => (
                  <div
                    key={v.id}
                    className="flex-1 p-4 bg-gray-100 rounded text-center"
                  >
                    <p className="text-sm text-gray-600 capitalize">
                      {v.estado}
                    </p>
                    <p className="text-2xl font-semibold">
                      {v.cantidad}
                    </p>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* Horas Trabajadas (Hoy) */}
          <section className="bg-white rounded-lg shadow p-4">
            <h3 className="text-xl font-medium mb-2">
              Horas Trabajadas (Hoy)
            </h3>

            {/* selector de tarifa */}
            <div className="mb-4 flex items-center gap-2">
              <label className="font-medium">Tarifa €/h:</label>
              <input
                type="number"
                className="w-24 p-1 border rounded"
                value={tarifa}
                onChange={(e) =>
                  setTarifa(parseFloat(e.target.value) || 0)
                }
              />
            </div>

            {loadingHoras ? (
              <p>Cargando horas…</p>
            ) : errorHoras ? (
              <p className="text-red-600">{errorHoras}</p>
            ) : (
              <div className="overflow-x-auto">
                <DataTable data={horasHoy} columns={horasHoyColumns} />
              </div>
            )}
          </section>

          {/* Vacaciones Solicitadas */}
          <section className="mb-8">
            <h2 className="text-2xl font-semibold mb-4">
              Vacaciones Solicitadas
            </h2>
            {loadingVac ? (
              <p>Cargando vacaciones…</p>
            ) : errorVac ? (
              <p className="text-red-600">{errorVac}</p>
            ) : (
              <div className="overflow-x-auto">
                <DataTable
                  data={vacaciones}
                  columns={[
                    { header: "ID", accessorKey: "id" },
                    { header: "Empleado", accessorKey: "empleado" },
                    {
                      header: "Fecha Inicio",
                      accessorKey: "fechaInicio",
                    },
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
                                  if (!res.ok)
                                    throw new Error("No se pudo aprobar");
                                  const upd = await res.json();
                                  setVacaciones((prev) =>
                                    prev.map((x) =>
                                      x.id === upd.id
                                        ? {
                                            ...x,
                                            estado:
                                              (upd.estado || "").toLowerCase(),
                                          }
                                        : x
                                    )
                                  );
                                  toast.success("Vacación aprobada");
                                } catch (e) {
                                  toast.error(e.message);
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
                                  if (!res.ok)
                                    throw new Error("No se pudo rechazar");
                                  const upd = await res.json();
                                  setVacaciones((prev) =>
                                    prev.map((x) =>
                                      x.id === upd.id
                                        ? {
                                            ...x,
                                            estado:
                                              (upd.estado || "").toLowerCase(),
                                          }
                                        : x
                                    )
                                  );
                                  toast.success("Vacación rechazada");
                                } catch (e) {
                                  toast.error(e.message);
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

          {/* Formularios Supervisor */}
          <div className="space-y-6 mt-6 w-full">
            <CrearEmpleadoForm />
            <EliminarEmpleadoForm />
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
