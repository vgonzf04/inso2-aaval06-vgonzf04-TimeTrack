// frontend/app/dashboard/page.js
"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";

import { AppSidebar } from "@/components/app-sidebar";
import { DataTable } from "@/components/data-table";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

import { TimerButton } from "@/components/timer-button";

export default function DashboardPage() {
  const [role, setRole] = useState(null);
  const router = useRouter();

  const [timecards, setTimecards] = useState([]);
  const [vacations, setVacations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    async function loadData() {
      try {
        // 1) Get my role
        const meRes = await fetch("http://localhost:3000/auth/me", {
          credentials: "include",
        });
        if (!meRes.ok) {
          router.push("/login");
          return;
        }
        const { role } = await meRes.json();
        setRole(role);

        // 2) Load timecards (sin cambios) …
        const tcRes = await fetch("http://localhost:3000/timecards", {
          credentials: "include",
        });
        if (!tcRes.ok) {
          router.push("/login");
          return;
        }
        const tcData = await tcRes.json();
        setTimecards(
          tcData.map((t) => ({
            id:       t.id,
            employee: t.employee?.name || "—",
            start:    t.start    || "—",
            end:      t.end      || "—",
            lat:      t.lat      ?? "—",
            lng:      t.lng      ?? "—",
            location: t.location || "—",
          }))
        );

        // 3) Load VACATIONS using el modelo JSON correcto:
        const vacRes = await fetch("http://localhost:3000/vacations", {
          credentials: "include",
        });
        if (!vacRes.ok) {
          router.push("/login");
          return;
        }
        const vacData = await vacRes.json();
        setVacations(
          vacData.map((v) => ({
            id:       v.id,
            employee: v.employee?.name           || "—",
            // Estas propiedades vienen de tu modelo:
            start:    v.start_date_str          || v.start_date?.slice(0, 10) || "—",
            end:      v.end_date_str            || v.end_date?.slice(0, 10)   || "—",
            status:   v.status                  || "—",
          }))
        );

        setLoading(false);
      } catch (e) {
        console.error(e);
        setError("Something went wrong");
        setLoading(false);
      }
    }

    loadData();
  }, [router]);

  // Callbacks sin cambios…
  function handleTimecardCreated(raw) {
    const entry = {
      id:       raw.id,
      employee: raw.employee?.name || "—",
      start:    raw.start    || "—",
      end:      raw.end      || "—",
      lat:      raw.lat      ?? "—",
      lng:      raw.lng      ?? "—",
      location: raw.location || "—",
    };
    setTimecards((prev) => [...prev, entry]);
    toast.success("Timecard added");
  }

  function handleTimecardClosed(raw) {
    setTimecards((prev) =>
      prev.map((t) =>
        t.id === raw.id ? { ...t, end: raw.end || "—" } : t
      )
    );
    toast.success("Timecard closed");
  }

  if (loading) {
    return (
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <div className="flex items-center justify-center h-screen">
            <p>Loading dashboard…</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }
  if (error) {
    return (
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <p className="text-red-600">Error: {error}</p>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const timecardColumns = [
    { header: "ID",       accessorKey: "id" },
    { header: "Employee", accessorKey: "employee" },
    { header: "Start",    accessorKey: "start" },
    { header: "End",      accessorKey: "end" },
    { header: "Lat",      accessorKey: "lat" },
    { header: "Lng",      accessorKey: "lng" },
    { header: "Location", accessorKey: "location" },
  ];
  const vacationColumns = [
    { header: "ID",       accessorKey: "id" },
    { header: "Employee", accessorKey: "employee" },
    { header: "Start",    accessorKey: "start" },
    { header: "End",      accessorKey: "end" },
    { header: "Status",   accessorKey: "status" },
  ];

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="bg-background sticky top-0 flex h-16 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Dashboard</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          {role === "supervisor" && (
            <Button size="sm" variant="outline" onClick={() => router.push("/admin")} className="ml-auto">
              Go to Admin
            </Button>
          )}
          <TimerButton
            onTimecardCreated={handleTimecardCreated}
            onTimecardClosed={handleTimecardClosed}
          />
        </header>

        <div className="flex flex-1 flex-col gap-4 p-4">
          {/* Timecards */}
          <section className="mb-8">
            <h2 className="text-2xl font-semibold mb-4">Timecards</h2>
            <div className="overflow-x-auto">
              <DataTable data={timecards} columns={timecardColumns} />
            </div>
          </section>

          {/* Vacations */}
          <section>
            <h2 className="text-2xl font-semibold mb-4">Vacation Requests</h2>
            <div className="overflow-x-auto">
              <DataTable data={vacations} columns={vacationColumns} />
            </div>
          </section>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
