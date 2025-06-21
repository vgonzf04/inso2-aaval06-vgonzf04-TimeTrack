"use client"

import { useEffect, useState } from "react"

export default function EliminarEmpleadoForm() {
  const [empleados, setEmpleados] = useState([])
  const [empleadoID, setEmpleadoID] = useState("")
  const [mensaje, setMensaje] = useState("")

  useEffect(() => {
   fetch("http://localhost:3000/empleados", {
     method: "GET",
     credentials: "include"
   })
     .then(res => res.json())
     .then(data => {
       setEmpleados(data)
     })
     .catch(err => {
       console.error("Error al cargar empleados:", err)
     })

  }, [])

  const handleSubmit = async (e) => {
    e.preventDefault()
    setMensaje("")

    if (!empleadoID) {
      setMensaje("Selecciona un empleado válido.")
      return
    }

    const confirmacion = window.confirm("¿Estás seguro de que quieres eliminar este empleado?")
    if (!confirmacion) return

    try {
      const response = await fetch(`http://localhost:3000/empleados/${empleadoID}`, {
        method: "DELETE",
        credentials: "include"
      })

      if (response.status === 204) {
        setMensaje("Empleado eliminado correctamente.")
        setEmpleados(empleados.filter(emp => emp.id !== parseInt(empleadoID)))
        setEmpleadoID("")
      } else {
        const error = await response.text()
        console.error("Error al eliminar:", error)
        setMensaje("Error al eliminar el empleado.")
      }
    } catch (err) {
      console.error(err)
      setMensaje("Error al eliminar el empleado.")
    }
  }
  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4 w-full mt-8">
      <label className="font-semibold">Selecciona un empleado a eliminar:</label>
      <select
        value={empleadoID}
        onChange={(e) => setEmpleadoID(e.target.value)}
        required
        className="border rounded p-2"
      >
        <option value="">-- Selecciona empleado --</option>
        {empleados.map(emp => (
          <option key={emp.id} value={emp.id}>
            {emp.nombre} (ID: {emp.id})
          </option>
        ))}
      </select>

      <button type="submit" className="bg-red-600 text-white px-4 py-2 rounded">
        Eliminar empleado
      </button>
      {mensaje && <p className="text-sm">{mensaje}</p>}
    </form>
  )
}
