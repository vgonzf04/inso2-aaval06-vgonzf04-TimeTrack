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

import { TimerButton } from "@/components/timer-button"
import { toast } from "sonner"

export default function Page() {
  const router = useRouter()

  const [fichajes, setFichajes] = useState([])
  const [vacaciones, setVacaciones] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // 1. Al montar, cargamos de golpe fichajes y vacaciones
  useEffect(() => {
    async function fetchData() {
      try {
        // --- CARGAR FICHAJES ---
        const resF = await fetch("http://localhost:3000/fichajes", {
          method: "GET",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
        })
        if (resF.status === 401 || resF.status === 403) {
          setLoading(false)
          router.push("/login")
          return
        }
        if (!resF.ok) throw new Error("Error al cargar fichajes")
        const dataF = await resF.json()
        // Aplanamos
        const fichajesFlat = dataF.map((f) => ({
          id: f.id,
          empleado: f.empleado?.nombre || "—",
          entrada: f.entrada || "—",
          salida: f.salida || "—",
          latitud: f.latitud ?? "—",
          longitud: f.longitud ?? "—",
          ubicacion: f.ubicacion || "—",
          // podrías añadir más campos si tu JSON los trae
        }))
        setFichajes(fichajesFlat)

        // --- CARGAR VACACIONES ---
        const resV = await fetch("http://localhost:3000/vacaciones", {
          method: "GET",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
        })
        if (resV.status === 401 || resV.status === 403) {
          setLoading(false)
          router.push("/login")
          return
        }
        if (!resV.ok) throw new Error("Error al cargar vacaciones")
        const dataV = await resV.json()
        const vacasFlat = dataV.map((v) => ({
          id: v.id,
          empleado: v.empleado?.nombre || "—",
          fechaInicio: v.Inicio ?? v.inicio ?? "—",
          fechaFin: v.Fin ?? v.fin ?? "—",
          estado: v.Estado ?? v.estado ?? "—",
        }))
        setVacaciones(vacasFlat)

        setLoading(false)
      } catch (err) {
        console.error(err)
        setError(err.message || "Algo salió mal")
        setLoading(false)
      }
    }
    fetchData()
  }, [router])

  // 2. Callback para añadir un fichaje nuevo a la tabla
  function handleFichajeCreado(nuevoFichajeCrudo) {
    // “Aplanamos” el objeto crudo exactamente igual que en fetchData:
    const fichajeFlat = {
      id: nuevoFichajeCrudo.id,
      empleado: nuevoFichajeCrudo.empleado?.nombre || "—",
      entrada: nuevoFichajeCrudo.entrada || "—",
      salida: nuevoFichajeCrudo.salida || "—",
      latitud: nuevoFichajeCrudo.latitud ?? "—",
      longitud: nuevoFichajeCrudo.longitud ?? "—",
      ubicacion: nuevoFichajeCrudo.ubicacion || "—",
    }
    // Lo añadimos al final del array fichajes
    setFichajes((prev) => [...prev, fichajeFlat])
    toast.success("Fichaje añadido a la tabla")
  }

  // 3. (Opcional) Callback para fichaje cerrado
  function handleFichajeCerrado(fichajeCerradoCrudo) {
    // Si quieres reflejar en la UI algo al cerrar el fichaje,
    // por ejemplo cambiar el estado de esa fila o moverlo a otra tabla,
    // puedes manejarlo aquí.
    toast.success("Fichaje cerrado correctamente")
    // En este ejemplo no modificamos el array de fichajes abiertos,
    // pero podrías, p.ej., volver a llamar a fetchData() o filtrar.
  }

  if (loading) {
    return (
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <div className="flex items-center justify-center h-screen w-full">
            <p className="text-lg">Cargando datos del dashboard...</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    )
  }

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

  // 4. Definimos las columnas planas para DataTable
  const fichajesColumns = [
    { header: "ID", accessorKey: "id" },
    { header: "Empleado", accessorKey: "empleado" },
    { header: "Entrada", accessorKey: "entrada" },
    { header: "Salida", accessorKey: "salida" },

    { header: "Latitud", accessorKey: "latitud" },
    { header: "Longitud", accessorKey: "longitud" },
    { header: "Ubicación", accessorKey: "ubicacion" },
  ]

  const vacacionesColumns = [
    { header: "ID", accessorKey: "id" },
    { header: "Empleado", accessorKey: "empleado" },
    { header: "Fecha Inicio", accessorKey: "fechaInicio" },
    { header: "Fecha Fin", accessorKey: "fechaFin" },
    { header: "Estado", accessorKey: "estado" },
  ]

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="bg-background sticky top-0 flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-[orientation=vertical]:h-4"
          />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Dashboard</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          {/* Pasamos ambos callbacks a TimerButton */}
          <TimerButton
            onFichajeCreado={handleFichajeCreado}
            onFichajeCerrado={handleFichajeCerrado}
          />
        </header>

        <div className="flex flex-1 flex-col gap-4 p-4">
          {/* DataTable de Fichajes */}
          <section className="mb-8">
            <h2 className="text-2xl font-semibold mb-4">Listado de Fichajes</h2>
            <div className="overflow-x-auto">
              <DataTable data={fichajes} columns={fichajesColumns} />
            </div>
          </section>

          {/* DataTable de Vacaciones */}
          <section>
            <h2 className="text-2xl font-semibold mb-4">
              Vacaciones Solicitadas
            </h2>
            <div className="overflow-x-auto">
              <DataTable data={vacaciones} columns={vacacionesColumns} />
            </div>
          </section>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
