// app/dashboard/page.js

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
  const [rol, setRol] = useState(null)
  const router = useRouter()

  const [fichajes, setFichajes] = useState([])
  const [vacaciones, setVacaciones] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  // 1) Al montar, pedimos /me, luego fichajes y vacaciones
  useEffect(() => {
    async function pedirRolYDatos() {
      try {
        // a) Obtener rol
        const resMe = await fetch("http://localhost:3000/me", {
          method: "GET",
          credentials: "include",
        })
        if (!resMe.ok) {
          router.push("/login")
          return
        }
        const dataMe = await resMe.json()
        setRol(dataMe.rol) // "supervisor" | "empleado"

        // b) Cargar fichajes
        const resF = await fetch("http://localhost:3000/fichajes", {
          method: "GET",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
        })
        if (resF.status === 401 || resF.status === 403) {
          router.push("/login")
          return
        }
        if (!resF.ok) throw new Error("Error al cargar fichajes")
        const dataF = await resF.json()
        const fichajesFlat = dataF.map((f) => ({
          id: f.id,
          empleado: f.empleado?.nombre || "—",
          entrada: f.entrada || "—",
          salida: f.salida || "—",
          latitud: f.latitud ?? "—",
          longitud: f.longitud ?? "—",
          ubicacion: f.ubicacion || "—",
        }))
        setFichajes(fichajesFlat)

        // c) Cargar vacaciones
        const resV = await fetch("http://localhost:3000/vacaciones", {
          method: "GET",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
        })
        if (resV.status === 401 || resV.status === 403) {
          router.push("/login")
          return
        }
        if (!resV.ok) throw new Error("Error al cargar vacaciones")
        const dataV = await resV.json()
        const vacasFlat = dataV.map((v) => ({
          id: v.id,
          empleado: v.empleado?.nombre || "—",
          fechaInicio: v.inicioStr ?? v.Inicio ?? "—",
          fechaFin: v.finStr ?? v.Fin ?? "—",
          estado: v.estado || v.Estado || "—",
        }))
        setVacaciones(vacasFlat)

        setLoading(false)
      } catch (err) {
        console.error(err)
        setError(err.message || "Algo salió mal")
        setLoading(false)
      }
    }

    pedirRolYDatos()
  }, [router])

  // 2) Callback para fichaje creado
  function handleFichajeCreado(nuevoFichajeCrudo) {
    const fichajeFlat = {
      id: nuevoFichajeCrudo.id,
      empleado: nuevoFichajeCrudo.empleado?.nombre || "—",
      entrada: nuevoFichajeCrudo.entrada || "—",
      salida: nuevoFichajeCrudo.salida || "—",
      latitud: nuevoFichajeCrudo.latitud ?? "—",
      longitud: nuevoFichajeCrudo.longitud ?? "—",
      ubicacion: nuevoFichajeCrudo.ubicacion || "—",
    }
    setFichajes((prev) => [...prev, fichajeFlat])
    toast.success("Fichaje añadido a la tabla")
  }

  // 3) Callback para fichaje cerrado
  function handleFichajeCerrado(fichajeCerradoCrudo) {
    const fichajeFlat = {
      id: fichajeCerradoCrudo.id,
      salida: fichajeCerradoCrudo.salida || "—",
    }
    setFichajes((prev) =>
      prev.map((f) =>
        f.id === fichajeFlat.id ? { ...f, salida: fichajeFlat.salida } : f
      )
    )
    toast.success("Fichaje cerrado correctamente")
  }

  // 4) Callback para vacación creada (lo llamará DatePicker)
  function handleVacacionCreada(nuevaVacacionCruda) {
   const vacaFlat = {
      id: nuevaVacacionCruda.id,
      empleado: nuevaVacacionCruda.empleado?.nombre ?? "—",
      fechaInicio:
        nuevaVacacionCruda.inicioStr ??
        nuevaVacacionCruda.Inicio ??
        nuevaVacacionCruda.inicio ??
        "—",
      fechaFin:
        nuevaVacacionCruda.finStr ??
        nuevaVacacionCruda.Fin ??
        nuevaVacacionCruda.fin ??
        "—",
      estado:
        nuevaVacacionCruda.estado ??
        nuevaVacacionCruda.Estado ??
        "—",
    }

    setVacaciones((prev) => [...prev, vacaFlat])
    toast.success("Solicitud de vacación enviada")
  }

  if (loading) {
    return (
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <div className="flex items-center justify-center h-screen w-full">
            <p className="text-lg">Cargando datos del dashboard…</p>
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

          {rol === "supervisor" && (
            <button
              type="button"
              onClick={() => router.push("/admin")}
              className="ml-auto px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              Ir a Administración
            </button>
          )}

          <TimerButton
            onFichajeCreado={handleFichajeCreado}
            onFichajeCerrado={handleFichajeCerrado}
          />
        </header>

        <div className="flex flex-1 flex-col gap-4 p-4">
          {/* ──────── Tabla de Fichajes ──────── */}
          <section className="mb-8">
            <h2 className="text-2xl font-semibold mb-4">Listado de Fichajes</h2>
            <div className="overflow-x-auto">
              <DataTable data={fichajes} columns={fichajesColumns} />
            </div>
          </section>

          {/* ──────── Tabla de Vacaciones ──────── */}
          <section className="mb-8">
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
