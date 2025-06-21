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
        const meData = await meRes.json();
        setRole(meData.role);

        // 2) Load timecards
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
            id: t.id,
            employee: t.employee?.name || "—",
            checkIn: t.checkIn || "—",
            checkOut: t.checkOut || "—",
            latitude: t.latitude ?? "—",
            longitude: t.longitude ?? "—",
            location: t.location || "—",
          }))
        );

        // 3) Load vacations
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
            id: v.id,
            employee: v.employee?.name || "—",
            startDate: v.startStr || v.start?.slice(0, 10) || "—",
            endDate: v.endStr || v.end?.slice(0, 10) || "—",
            status: (v.status || v.estado || "—").toLowerCase(),
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

  // Callback when a new timecard is created
  function handleTimecardCreated(raw) {
    const entry = {
      id: raw.id,
      employee: raw.employee?.name || "—",
      checkIn: raw.checkIn || "—",
      checkOut: raw.checkOut || "—",
      latitude: raw.latitude ?? "—",
      longitude: raw.longitude ?? "—",
      location: raw.location || "—",
    };
    setTimecards((prev) => [...prev, entry]);
    toast.success("Timecard added");
  }

  // Callback when a timecard is closed
  function handleTimecardClosed(raw) {
    setTimecards((prev) =>
      prev.map((t) =>
        t.id === raw.id ? { ...t, checkOut: raw.checkOut || "—" } : t
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
    { header: "ID", accessorKey: "id" },
    { header: "Employee", accessorKey: "employee" },
    { header: "Check-In", accessorKey: "checkIn" },
    { header: "Check-Out", accessorKey: "checkOut" },
    { header: "Latitude", accessorKey: "latitude" },
    { header: "Longitude", accessorKey: "longitude" },
    { header: "Location", accessorKey: "location" },
  ];

  const vacationColumns = [
    { header: "ID", accessorKey: "id" },
    { header: "Employee", accessorKey: "employee" },
    { header: "Start Date", accessorKey: "startDate" },
    { header: "End Date", accessorKey: "endDate" },
    { header: "Status", accessorKey: "status" },
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
            <Button
              size="sm"
              variant="outline"
              onClick={() => router.push("/admin")}
              className="ml-auto"
            >
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
