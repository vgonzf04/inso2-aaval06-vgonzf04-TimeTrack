"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import EmpleadoForm from "@/components/EmpleadoForm"

export default function CrearEmpleadoPage() {
  const [rol, setRol] = useState(null)
  const router = useRouter()

  useEffect(() => {
    const userRol = localStorage.getItem("rol")
    setRol(userRol)

    // Redirigir si no es supervisor
    if (userRol && userRol !== "supervisor") {
      router.push("/admin")
    }
  }, [router])

  if (rol !== "supervisor") {
    return <p>Verificando permisos...</p>
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Añadir nuevo empleado</h1>
      <EmpleadoForm />
    </div>
  )
}
