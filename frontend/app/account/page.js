// frontend/app/account/page.js

"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"

import { AppSidebar } from "@/components/app-sidebar"
import { DataTable } from "@/components/data-table"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Button } from "@/components/ui/button"

export default function AccountPage() {
  const router = useRouter()

  // —— Perfil —— 
  const [perfil, setPerfil] = useState(null)
  const [loadingPerfil, setLoadingPerfil] = useState(true)
  const [errorPerfil, setErrorPerfil] = useState(null)

  // —— Horas trabajadas —— 
  const [horas, setHoras] = useState([])          
  const [loadingHoras, setLoadingHoras] = useState(true)
  const [errorHoras, setErrorHoras] = useState(null)

  // tarifa fija
  const tarifa = 15

  // Helper YYYY-MM-DD
  function formatYYYYMMDD(d) {
    const yyyy = d.getFullYear()
    const mm = String(d.getMonth() + 1).padStart(2, "0")
    const dd = String(d.getDate()).padStart(2, "0")
    return `${yyyy}-${mm}-${dd}`
  }

  useEffect(() => {
    async function fetchMiPerfil() {
      try {
        const res = await fetch("http://localhost:3000/empleados/me", {
          method: "GET",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
        })
        if (res.status === 401 || res.status === 403) {
          router.push("/login"); return
        }
        if (!res.ok) throw new Error("Error al cargar perfil")
        setPerfil(await res.json())
      } catch (err) {
        setErrorPerfil(err.message)
      } finally {
        setLoadingPerfil(false)
      }
    }
    fetchMiPerfil()
  }, [router])

  useEffect(() => {
    if (!perfil?.id) return
    async function fetchHoras() {
      try {
        const hoy = formatYYYYMMDD(new Date())
        const res = await fetch(
          `http://localhost:3000/dashboard/horas-periodo?inicio=${hoy}&fin=${hoy}&empleado_id=${perfil.id}`,
          { method: "GET", credentials: "include" }
        )
        if (res.status === 401 || res.status === 403) {
          router.push("/login"); return
        }
        if (!res.ok) throw new Error("Error al cargar horas")
        const data = await res.json()
        setHoras(
          (Array.isArray(data) ? data : []).map(x => ({
            id: x.empleado_id,
            empleado_id: x.empleado_id,
            nombre: x.nombre,
            total_horas: x.total_horas,
          }))
        )
      } catch (err) {
        setErrorHoras(err.message)
      } finally {
        setLoadingHoras(false)
      }
    }
    fetchHoras()
  }, [perfil, router])

  if (loadingPerfil) return (
    <SidebarProvider><AppSidebar /><SidebarInset>
      <div className="flex items-center justify-center h-screen"><p>Cargando…</p></div>
    </SidebarInset></SidebarProvider>
  )
  if (errorPerfil) return (
    <SidebarProvider><AppSidebar /><SidebarInset>
      <p className="text-red-600">Error: {errorPerfil}</p>
    </SidebarInset></SidebarProvider>
  )

  const filasPerfil = [{
    id: perfil.id,
    nombre: perfil.nombre,
    email: perfil.email,
    cargo: perfil.cargo,
    fechaContratacion: perfil.fecha_contratacion ?? perfil.FechaContratacion,
    supervisorID: perfil.supervisor_id ?? perfil.SupervisorID,
    rol: perfil.rol ?? perfil.Rol,
  }]
  const colsPerfil = [
    { header: "ID", accessorKey: "id" },
    { header: "Nombre", accessorKey: "nombre" },
    { header: "Email", accessorKey: "email" },
    { header: "Cargo", accessorKey: "cargo" },
    { header: "Fecha Contratación", accessorKey: "fechaContratacion" },
    { header: "Supervisor ID", accessorKey: "supervisorID" },
    { header: "Rol", accessorKey: "rol" },
  ]

  const colsHoras = [
    { header: "Empleado ID", accessorKey: "empleado_id" },
    { header: "Nombre", accessorKey: "nombre" },
    {
      header: "Total Horas",
      accessorKey: "total_horas",
      cell: ({ getValue }) => Number(getValue()).toFixed(3),
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
  ]

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="bg-background sticky top-0 flex h-16 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb><BreadcrumbList>
            <BreadcrumbItem><BreadcrumbPage>Mi Cuenta</BreadcrumbPage></BreadcrumbItem>
          </BreadcrumbList></Breadcrumb>
          <div className="ml-auto">
            <Button size="sm" variant="outline" onClick={() => router.push("/dashboard")}>
              ← Dashboard
            </Button>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4">
          <section>
            <h2 className="text-2xl font-semibold mb-4">Datos de mi usuario</h2>
            <DataTable data={filasPerfil} columns={colsPerfil} />
          </section>
          <section>
            <h2 className="text-2xl font-semibold mb-4">Horas Trabajadas Hoy</h2>
            {loadingHoras
              ? <p>Cargando horas…</p>
              : errorHoras
                ? <p className="text-red-600">Error: {errorHoras}</p>
                : <DataTable data={horas} columns={colsHoras} />
            }
          </section>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
