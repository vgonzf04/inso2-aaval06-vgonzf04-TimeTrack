import React, { useState } from "react"

export default function EmpleadoForm() {
  const [formData, setFormData] = useState({
    nombre: "",
    email: "",
    cargo: "",
    supervisor_id: "", // puede ser vacío o null
    rol: "empleado"
  })

  const [mensaje, setMensaje] = useState("")

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({ ...prev, [name]: value }))
  }

  const handleSubmit = (e) => {
    e.preventDefault()

    const payload = {
      ...formData,
      supervisor_id: formData.supervisor_id ? parseInt(formData.supervisor_id) : null
    }

    console.log("Payload enviado:", payload)

    fetch("http://localhost:3000/empleados/", {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "include",
      body: JSON.stringify(payload)
    })

    .then(async response => {
      if (!response.ok) {
        const errorText = await response.text()
        console.error("Error del servidor:", errorText)
        throw new Error("Error al crear el empleado.")
      }

      let data
      try {
        data = await response.json()
      } catch (err) {
        console.warn("⚠️ La respuesta no contenía JSON válido.")
        data = null
      }

      setMensaje("Empleado creado correctamente.")
      setFormData({
        nombre: "",
        email: "",
        cargo: "",
        supervisor_id: "",
        rol: "empleado"
      })

      console.log("Respuesta del backend:", data)
    })
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4 w-full">
      <input name="nombre" placeholder="Nombre" value={formData.nombre} onChange={handleChange} required />
      <input name="email" type="email" placeholder="Email" value={formData.email} onChange={handleChange} required />
      <input name="cargo" placeholder="Cargo" value={formData.cargo} onChange={handleChange} required />
      <input name="supervisor_id" type="number" placeholder="ID del supervisor (opcional)" value={formData.supervisor_id} onChange={handleChange} />
      <select name="rol" value={formData.rol} onChange={handleChange}>
        <option value="empleado">Empleado</option>
        <option value="supervisor">Supervisor</option>
      </select>
      <button type="submit" className="bg-blue-600 text-white px-4 py-2 rounded">Crear empleado</button>
      {mensaje && <p>{mensaje}</p>}
    </form>
  )
}
