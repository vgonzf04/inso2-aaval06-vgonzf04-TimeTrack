// frontend/app/account/page.js

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

export default function AccountPage() {
  const router = useRouter();

  // —— Profile ——
  const [profile, setProfile] = useState(null);
  const [loadingProfile, setLoadingProfile] = useState(true);
  const [errorProfile, setErrorProfile] = useState(null);

  // —— Hours Worked ——
  const [hours, setHours] = useState([]);
  const [loadingHours, setLoadingHours] = useState(true);
  const [errorHours, setErrorHours] = useState(null);

  // Fixed rate €/h
  const rate = 15;

  // Helper to format a Date as "YYYY-MM-DD"
  function formatYYYYMMDD(d) {
    const yyyy = d.getFullYear();
    const mm = String(d.getMonth() + 1).padStart(2, "0");
    const dd = String(d.getDate()).padStart(2, "0");
    return `${yyyy}-${mm}-${dd}`;
  }

  // 1) Load my profile
  useEffect(() => {
    async function fetchProfile() {
      try {
        const res = await fetch("http://localhost:3000/employees/me", {
          method: "GET",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
        });
        if (res.status === 401 || res.status === 403) {
          router.push("/login");
          return;
        }
        if (!res.ok) throw new Error("Failed to load profile");
        setProfile(await res.json());
      } catch (err) {
        setErrorProfile(err.message);
      } finally {
        setLoadingProfile(false);
      }
    }
    fetchProfile();
  }, [router]);

  // 2) Once profile is loaded, fetch today's hours
  useEffect(() => {
    if (!profile?.id) return;

    async function fetchHours() {
      try {
        const today = formatYYYYMMDD(new Date());
        const res = await fetch(
          `http://localhost:3000/dashboard/hours-period?start=${today}&end=${today}&employee_id=${profile.id}`,
          { method: "GET", credentials: "include" }
        );
        if (res.status === 401 || res.status === 403) {
          router.push("/login");
          return;
        }
        if (!res.ok) throw new Error("Failed to load hours worked");
        const data = await res.json();
        setHours(
          (Array.isArray(data) ? data : []).map((x) => ({
            id: x.employee_id,
            employee_id: x.employee_id,
            name: x.name,
            total_hours: x.total_hours,
          }))
        );
      } catch (err) {
        setErrorHours(err.message);
      } finally {
        setLoadingHours(false);
      }
    }

    fetchHours();
  }, [profile, router]);

  // Loading & error states for profile
  if (loadingProfile) {
    return (
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <div className="flex items-center justify-center h-screen w-full">
            <p className="text-lg">Loading profile…</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }
  if (errorProfile) {
    return (
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <p className="text-red-600">Error: {errorProfile}</p>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  // Prepare profile table data
  const profileRows = [
    {
      id: profile.id,
      name: profile.nombre,
      email: profile.email,
      position: profile.cargo,
      hiringDate:
        profile.fecha_contratacion ?? profile.FechaContratacion ?? "—",
      supervisorId:
        profile.supervisor_id ?? profile.SupervisorID ?? "—",
      role: profile.rol ?? profile.Rol ?? "—",
    },
  ];
  const profileCols = [
    { header: "ID", accessorKey: "id" },
    { header: "Name", accessorKey: "name" },
    { header: "Email", accessorKey: "email" },
    { header: "Position", accessorKey: "position" },
    { header: "Hiring Date", accessorKey: "hiringDate" },
    { header: "Supervisor ID", accessorKey: "supervisorId" },
    { header: "Role", accessorKey: "role" },
  ];

  // Prepare hours table columns
  const hoursCols = [
    { header: "Employee ID", accessorKey: "employee_id" },
    { header: "Name", accessorKey: "name" },
    {
      header: "Total Hours",
      accessorKey: "total_hours",
      cell: ({ getValue }) => Number(getValue()).toFixed(3),
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
    <SidebarProvider>
      <AppSidebar />

      <SidebarInset>
        {/* ─── Header ─── */}
        <header className="bg-background sticky top-0 flex h-16 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>My Account</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <div className="ml-auto">
            <Button
              size="sm"
              variant="outline"
              onClick={() => router.push("/dashboard")}
            >
              ← Dashboard
            </Button>
          </div>
        </header>
        {/* ─── End Header ─── */}

        <div className="flex flex-1 flex-col gap-4 p-4">
          {/* ─── Profile Table ─── */}
          <section>
            <h2 className="text-2xl font-semibold mb-4">User Details</h2>
            <div className="overflow-x-auto">
              <DataTable data={profileRows} columns={profileCols} />
            </div>
          </section>

          {/* ─── Hours Worked Today ─── */}
          <section>
            <h2 className="text-2xl font-semibold mb-4">
              Hours Worked Today
            </h2>
            {loadingHours ? (
              <p>Loading hours…</p>
            ) : errorHours ? (
              <p className="text-red-600">Error: {errorHours}</p>
            ) : (
              <div className="overflow-x-auto">
                <DataTable data={hours} columns={hoursCols} />
              </div>
            )}
          </section>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
