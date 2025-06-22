// frontend/app/admin/page.js
"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";

import { AppSidebar } from "@/components/app-sidebar";
import { DataTable } from "@/components/data-table";
import { SiteHeader } from "@/components/site-header";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

import CreateEmployeeForm from "@/components/employee";
import DeleteEmployeeForm from "@/components/employee-delete";

// Helper para formatear fecha a YYYY-MM-DD
function formatYYYYMMDD(d) {
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd}`;
}

export default function AdminPage() {
  const router = useRouter();

  // ── Solicitudes de vacaciones ──
  const [vacations, setVacations] = useState([]);
  const [loadingVac, setLoadingVac] = useState(true);
  const [errorVac, setErrorVac] = useState(null);

  // ── Estadísticas de fichajes ──
  const [openCount, setOpenCount] = useState(0);
  const [closedCount, setClosedCount] = useState(0);
  const [loadingChecks, setLoadingChecks] = useState(true);
  const [errorChecks, setErrorChecks] = useState(null);

  // ── Vacaciones por estado ──
  const [vacByStatus, setVacByStatus] = useState([]);
  const [loadingVacStatus, setLoadingVacStatus] = useState(true);
  const [errorVacStatus, setErrorVacStatus] = useState(null);

  // ── Horas trabajadas hoy ──
  const [hoursToday, setHoursToday] = useState([]);
  const [loadingHours, setLoadingHours] = useState(true);
  const [errorHours, setErrorHours] = useState(null);

  // ── Tarifa por hora ──
  const [rate, setRate] = useState(15);

  // 1) Cargar todas las solicitudes de vacaciones de los empleados a mi cargo
  useEffect(() => {
    async function loadVacations() {
      try {
        const me = await fetch("http://localhost:3000/auth/me", { credentials: "include" });
        if (!me.ok) return router.push("/login");

        const res = await fetch("http://localhost:3000/vacations/employees", {
          credentials: "include",
        });
        if (!res.ok) {
          if (res.status === 401 || res.status === 403) {
            setLoadingVac(false);
            return router.push("/login");
          }
          throw new Error();
        }

        const data = await res.json();
        // Mapeamos usando start_date_str / end_date_str
        setVacations(
          data.map((v) => ({
            id:        v.id,
            employee:  v.employee?.name  || "—",
            startDate: v.start_date_str  || v.start_date?.slice(0, 10) || "—",
            endDate:   v.end_date_str    || v.end_date?.slice(0, 10)   || "—",
            status:    v.status          || "—",
          }))
        );
      } catch (e) {
        console.error(e);
        setErrorVac("Error loading vacations");
      } finally {
        setLoadingVac(false);
      }
    }
    loadVacations();
  }, [router]);

  // 2) Cargar estadísticas (fichajes, vacations by status, horas)
  useEffect(() => {
    const today = formatYYYYMMDD(new Date());

    // A) Fichajes abiertos/cerrados hoy
    (async () => {
      try {
        const res = await fetch(
          `http://localhost:3000/dashboard/checkins-day?day=${today}`,
          { credentials: "include" }
        );
        if (!res.ok) throw new Error();
        const { punches_open, punches_closed } = await res.json();
        setOpenCount(punches_open ?? 0);
        setClosedCount(punches_closed ?? 0);
      } catch {
        setErrorChecks("Error loading check-ins");
      } finally {
        setLoadingChecks(false);
      }
    })();

    // B) Vacaciones por estado
    (async () => {
      try {
        const res = await fetch(
          "http://localhost:3000/dashboard/vacations-by-status",
          { credentials: "include" }
        );
        if (!res.ok) throw new Error();
        const data = await res.json();
        setVacByStatus(
          data.map((x) => ({
            id:    x.state,
            state: x.state,
            count: x.quantity,
          }))
        );
      } catch {
        setErrorVacStatus("Error loading vacation statuses");
      } finally {
        setLoadingVacStatus(false);
      }
    })();

    // C) Horas trabajadas hoy (todos los empleados)
    (async () => {
      try {
        const res = await fetch(
          `http://localhost:3000/dashboard/hours-period?start=${today}&end=${today}`,
          { credentials: "include" }
        );
        if (!res.ok) throw new Error();
        const data = await res.json();
        setHoursToday(
          data.map((x) => ({
            id:            x.employee_id,
            employee_id:   x.employee_id,
            name:          x.name,
            total_hours:   x.total_hours,
            total_minutes: x.total_minutes,
          }))
        );
      } catch {
        setErrorHours("Error loading hours");
      } finally {
        setLoadingHours(false);
      }
    })();
  }, [router]);

  const hoursColumns = [
    { header: "Employee ID", accessorKey: "employee_id" },
    { header: "Name",        accessorKey: "name"        },
    {
      header: "Total Hours",
      accessorKey: "total_hours",
      cell: ({ getValue }) => Number(getValue() ?? 0).toFixed(3),
    },
    {
      header: "Minutes",
      accessorKey: "total_minutes",
      cell: ({ getValue }) => Number(getValue() ?? 0).toFixed(2),
    },
    {
      header: "€/h",
      accessorKey: "rate",
      cell: () => rate.toFixed(2),
    },
    {
      header: "Total €",
      accessorKey: "total_hours",
      cell: ({ getValue }) => (getValue() * rate).toFixed(2),
    },
  ];

  return (
    <SidebarProvider
      style={{
        "--sidebar-width":  "calc(var(--spacing) * 72)",
        "--header-height":  "calc(var(--spacing) * 12)",
      }}
    >
      <AppSidebar variant="inset" />
      <SidebarInset>
        <SiteHeader />

        {/* Check-Ins Today */}
        <section className="bg-white rounded-lg shadow p-4 mb-4">
          <h3 className="text-xl font-medium mb-2">Check-Ins (Today)</h3>
          {loadingChecks
            ? <p>Loading check-ins…</p>
            : errorChecks
              ? <p className="text-red-600">{errorChecks}</p>
              : (
                <div className="flex gap-6">
                  <div className="flex-1 p-4 bg-gray-100 rounded text-center">
                    <p className="text-sm text-gray-600">Open</p>
                    <p className="text-2xl font-semibold">{openCount}</p>
                  </div>
                  <div className="flex-1 p-4 bg-gray-100 rounded text-center">
                    <p className="text-sm text-gray-600">Closed</p>
                    <p className="text-2xl font-semibold">{closedCount}</p>
                  </div>
                </div>
              )
          }
        </section>

        {/* Vacations by Status */}
        <section className="bg-white rounded-lg shadow p-4 mb-4">
          <h3 className="text-xl font-medium mb-2">Vacations by Status</h3>
          {loadingVacStatus
            ? <p>Loading…</p>
            : errorVacStatus
              ? <p className="text-red-600">{errorVacStatus}</p>
              : (
                <div className="flex gap-6">
                  {vacByStatus.map((v) => (
                    <div key={v.id} className="flex-1 p-4 bg-gray-100 rounded text-center">
                      <p className="text-sm text-gray-600 capitalize">{v.state}</p>
                      <p className="text-2xl font-semibold">{v.count}</p>
                    </div>
                  ))}
                </div>
              )
          }
        </section>

        {/* Hours Worked (Today) */}
        <section className="bg-white rounded-lg shadow p-4 mb-4">
          <h3 className="text-xl font-medium mb-2">Hours Worked (Today)</h3>
          <div className="mb-4 flex items-center gap-2">
            <label className="font-medium">Hourly Rate (€):</label>
            <input
              type="number"
              className="w-24 p-1 border rounded"
              value={rate}
              onChange={(e) => setRate(parseFloat(e.target.value) || 0)}
            />
          </div>
          {loadingHours
            ? <p>Loading hours…</p>
            : errorHours
              ? <p className="text-red-600">{errorHours}</p>
              : (
                <div className="overflow-x-auto">
                  <DataTable data={hoursToday} columns={hoursColumns} />
                </div>
              )
          }
        </section>

        {/* Vacation Requests Table */}
        <section className="mb-8">
          <h2 className="text-2xl font-semibold mb-4">Vacation Requests</h2>
          {loadingVac
            ? <p>Loading vacations…</p>
            : errorVac
              ? <p className="text-red-600">{errorVac}</p>
              : (
                <div className="overflow-x-auto">
                  <DataTable
                    data={vacations}
                    columns={[
                      { header: "ID",          accessorKey: "id" },
                      { header: "Employee",    accessorKey: "employee" },
                      { header: "Start Date",  accessorKey: "startDate" },
                      { header: "End Date",    accessorKey: "endDate" },
                      { header: "Status",      accessorKey: "status" },
                      {
                        id:     "actions",
                        header: "Actions",
                        cell: ({ row }) => {
                          const v = row.original;
                          if (v.status !== "pending") return null;
                          return (
                            <div className="flex gap-2">
                              <Button
                                size="sm"
                                variant="outline"
                                onClick={async () => {
                                  try {
                                    const res = await fetch(
                                      `http://localhost:3000/vacations/${v.id}/approve`,
                                      { method: "PUT", credentials: "include" }
                                    );
                                    if (!res.ok) throw new Error();
                                    const updated = await res.json();
                                    // Actualizamos sólo el estado en la tabla
                                    setVacations((prev) =>
                                      prev.map((x) =>
                                        x.id === updated.id
                                          ? { ...x, status: updated.status }
                                          : x
                                      )
                                    );
                                    toast.success("Vacation approved");
                                  } catch {
                                    toast.error("Failed to approve");
                                  }
                                }}
                              >
                                Approve
                              </Button>
                              <Button
                                size="sm"
                                variant="destructive"
                                onClick={async () => {
                                  try {
                                    const res = await fetch(
                                      `http://localhost:3000/vacations/${v.id}/reject`,
                                      { method: "PUT", credentials: "include" }
                                    );
                                    if (!res.ok) throw new Error();
                                    const updated = await res.json();
                                    setVacations((prev) =>
                                      prev.map((x) =>
                                        x.id === updated.id
                                          ? { ...x, status: updated.status }
                                          : x
                                      )
                                    );
                                    toast.success("Vacation rejected");
                                  } catch {
                                    toast.error("Failed to reject");
                                  }
                                }}
                              >
                                Reject
                              </Button>
                            </div>
                          );
                        },
                      },
                    ]}
                  />
                </div>
              )
          }
        </section>

        {/* Formularios de supervisor */}
        <div className="space-y-6 mt-6 w-full">
          <CreateEmployeeForm />
          <DeleteEmployeeForm />
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
