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
import { Button } from "@/components/ui/button"   // ← Importamos el componente Button

export default function AccountPage() {
  const router = useRouter()

  const [perfil, setPerfil] = useState(null)    // Aquí guardamos el objeto Empleado
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    async function fetchMiPerfil() {
      try {
        const res = await fetch("http://localhost:3000/empleados/me", {
          method: "GET",
          credentials: "include",       // <- enviamos la cookie JWT
          headers: { "Content-Type": "application/json" },
        })

        // Si no está autenticado o token inválido, redirigimos a /login
        if (res.status === 401 || res.status === 403) {
          router.push("/login")
          return
        }
        if (!res.ok) {
          throw new Error("Error al cargar perfil de empleado")
        }

        const data = await res.json()
        // data debe ser objeto con { id, nombre, email, cargo, fecha_contratacion, supervisor_id, rol }
        setPerfil(data)
        setLoading(false)
      } catch (err) {
        console.error("fetchMiPerfil:", err)
        setError(err.message || "Algo salió mal al cargar el perfil")
        setLoading(false)
      }
    }

    fetchMiPerfil()
  }, [router])

  // 1) Mientras cargamos, mostramos un spinner / loading
  if (loading) {
    return (
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <div className="flex items-center justify-center h-screen w-full">
            <p className="text-lg">Cargando datos de mi cuenta…</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    )
  }

  // 2) Si dio error al traer el perfil, lo mostramos
  if (error) {
    return (
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <div className="p-4">
            <p className="text-red-600">Error: {error}</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    )
  }

  // 3) Verificamos que ‘perfil’ exista y tenga un 'id' numérico
  if (!perfil || typeof perfil.id !== "number") {
    return (
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <div className="p-4">
            <p className="text-red-600">No se encontró perfil de usuario.</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    )
  }

  // 4) Construimos el array de “filas” para pasarlo a DataTable.
  //    Éste es un array de un solo elemento a partir de ‘perfil’
  const filas = [
    {
      id: perfil.id,                                      // ← debe ser un número
      nombre: perfil.nombre ?? "—",
      email: perfil.email ?? "—",
      cargo: perfil.cargo ?? "—",
      // Ajusta si en tu modelo Go se llama `fecha_contratacion` o `FechaContratacion`
      fechaContratacion: perfil.fecha_contratacion ?? perfil.FechaContratacion ?? "—",
      // `supervisor_id` en Go es *uint o nil, en JS vendrá null o número
      supervisorID: perfil.supervisor_id ?? perfil.SupervisorID ?? "—",
      rol: perfil.rol ?? perfil.Rol ?? "—",
    },
  ]

  // 5) Definimos las columnas de la tabla; las claves EXACTAS han de coincidir con los campos de ‘filas’
  const columnas = [
    { header: "ID", accessorKey: "id" },
    { header: "Nombre", accessorKey: "nombre" },
    { header: "Email", accessorKey: "email" },
    { header: "Cargo", accessorKey: "cargo" },
    { header: "Fecha Contratación", accessorKey: "fechaContratacion" },
    { header: "Supervisor ID", accessorKey: "supervisorID" },
    { header: "Rol", accessorKey: "rol" },
  ]

  return (
    <SidebarProvider>
      <AppSidebar />

      <SidebarInset>
        {/* ─── Header ─── */}
        <header className="bg-background sticky top-0 flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-[orientation=vertical]:h-4"
          />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Mi Cuenta</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>

          {/* ── Botón “← Dashboard” (aparece en la parte superior derecha del header) ── */}
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
        {/* ─── Fin del Header ─── */}

        {/* ─── Tabla con datos de mi usuario ─── */}
        <div className="flex flex-1 flex-col gap-4 p-4">
          <section>
            <h2 className="text-2xl font-semibold mb-4">Datos de mi usuario</h2>
            <div className="overflow-x-auto">
              <DataTable data={filas} columns={columnas} />
            </div>
          </section>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
