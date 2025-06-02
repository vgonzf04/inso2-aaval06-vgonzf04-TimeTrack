"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"

import { AppSidebar } from "@/components/app-sidebar"
import { DataTable } from "@/components/data-table"
import { SectionCards } from "@/components/section-cards"
import { SiteHeader } from "@/components/site-header"
import {
  SidebarInset,
  SidebarProvider,
} from "@/components/ui/sidebar"
import { Button } from "@/components/ui/button"
import { toast } from "sonner"

import CrearEmpleadoForm from "@/components/empleado"
import EliminarEmpleadoForm from "@/components/eliminarEmpleado"

export default function Page() {
  const router = useRouter()

  const [vacaciones, setVacaciones] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    async function fetchVacacionesEmpleados() {
      try {
        const res = await fetch("http://localhost:3000/vacaciones/empleados", {
          method: "GET",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
        })
        if (res.status === 401 || res.status === 403) {
          setLoading(false)
          router.push("/login")
          return
        }
        if (!res.ok) throw new Error("Error al cargar vacaciones de empleados")
        const data = await res.json()

        const vacasFlat = data.map((v) => ({
          id: v.id,
          empleado: v.empleado?.nombre || "—",
          fechaInicio: v.inicioStr || "—",
          fechaFin: v.finStr || "—",
          // Ojo: asegurarnos de que 'estado' venga en minúsculas ("pendiente", "aprobada", "rechazada")
          estado: (v.estado || "").toLowerCase(),
        }))

        setVacaciones(vacasFlat)
        setLoading(false)
      } catch (err) {
        console.error(err)
        setError(err.message || "Algo salió mal")
        setLoading(false)
      }
    }

    fetchVacacionesEmpleados()
  }, [router])

  // Función para aprobar una vacación
  async function handleAprobar(id) {
    try {
      const res = await fetch(`http://localhost:3000/vacaciones/${id}/aprobar`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
      })
      if (!res.ok) throw new Error("Error al aprobar la vacación")
      const actualizado = await res.json()
      // Actualizar el campo 'estado' en el array
      setVacaciones((prev) =>
        prev.map((v) =>
          v.id === actualizado.id
            ? { ...v, estado: (actualizado.estado || "").toLowerCase() }
            : v
        )
      )
      console.log(actualizado.estado)
      console.log(vacaciones)
      toast.success("Vacación aprobada")
    } catch (err) {
      console.error(err)
      toast.error(err.message || "No se pudo aprobar la vacación")
    }
  }

  // Función para rechazar una vacación
  async function handleRechazar(id) {
    try {
      const res = await fetch(`http://localhost:3000/vacaciones/${id}/rechazar`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
      })
      if (!res.ok) throw new Error("Error al rechazar la vacación")
      const actualizado = await res.json()
      setVacaciones((prev) =>
        prev.map((v) =>
          v.id === actualizado.id
            ? { ...v, estado: (actualizado.estado || "").toLowerCase() }
            : v
        )
      )
      toast.success("Vacación rechazada")
    } catch (err) {
      console.error(err)
      toast.error(err.message || "No se pudo rechazar la vacación")
    }
  }

  const vacacionesColumns = [
    { header: "ID", accessorKey: "id" },
    { header: "Empleado", accessorKey: "empleado" },
    { header: "Fecha Inicio", accessorKey: "fechaInicio" },
    { header: "Fecha Fin", accessorKey: "fechaFin" },
    { header: "Estado", accessorKey: "estado" },
    {
      id: "acciones",
      header: "Acciones",
      cell: ({ row }) => {
        const v = row.original
        // Solo mostramos botones si el estado es exactamente "pendiente"
        if (v.estado !== "pendiente") {
          return null
        }
        return (
          <div className="flex gap-2">
            <Button
              size="sm"
              variant="outline"
              onClick={() => handleAprobar(v.id)}
            >
              Aprobar
            </Button>
            <Button
              size="sm"
              variant="destructive"
              onClick={() => handleRechazar(v.id)}
            >
              Rechazar
            </Button>
          </div>
        )
      },
    },
  ]

  if (loading) {
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
          <div className="flex items-center justify-center h-64">
            <p className="text-lg">Cargando vacaciones de empleados…</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    )
  }

  if (error) {
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
          <div className="p-4">
            <p className="text-red-600">Error: {error}</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    )
  }

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
          <SectionCards />

          <div className="overflow-x-auto">
            <DataTable data={vacaciones} columns={vacacionesColumns} />
          </div>

          {/* Formularios solo para supervisores */}
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
  )
}
