// frontend/components/employee-form.jsx
"use client";

import React, { useState } from "react";
import { toast } from "sonner";

export default function EmployeeForm({ onCreated }) {
  const [formData, setFormData] = useState({
    name: "",
    email: "",
    position: "",
    supervisor_id: "",  // letal si está vacío
    role: "employee",   // o "supervisor"
  });

  // Maneja cambios en inputs
  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  // Submit al backend
  const handleSubmit = async (e) => {
    e.preventDefault();

    const payload = {
      name: formData.name,
      email: formData.email,
      position: formData.position,
      supervisor_id: formData.supervisor_id
        ? parseInt(formData.supervisor_id, 10)
        : null,
      role: formData.role.trim().toLowerCase(),
    };

    try {
      const res = await fetch("http://localhost:3000/employees", {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        const errText = await res.text();
        console.error("Server error:", errText);
        throw new Error("Error al crear el empleado");
      }

      const newEmp = await res.json();
      toast.success("Empleado creado correctamente");
      setFormData({
        name: "",
        email: "",
        position: "",
        supervisor_id: "",
        role: "employee",
      });

      // Si el padre quiere actualizar la tabla:
      onCreated?.(newEmp);
    } catch (err) {
      console.error(err);
      toast.error(err.message);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="flex flex-col gap-4 w-full max-w-md"
    >
      <input
        name="name"
        placeholder="Nombre"
        value={formData.name}
        onChange={handleChange}
        required
      />
      <input
        name="email"
        type="email"
        placeholder="Email"
        value={formData.email}
        onChange={handleChange}
        required
      />
      <input
        name="position"
        placeholder="Cargo"
        value={formData.position}
        onChange={handleChange}
        required
      />
      <input
        name="supervisor_id"
        type="number"
        placeholder="ID Supervisor (opcional)"
        value={formData.supervisor_id}
        onChange={handleChange}
      />
      <select
        name="role"
        value={formData.role}
        onChange={handleChange}
      >
        <option value="employee">Empleado</option>
        <option value="supervisor">Supervisor</option>
      </select>
      <button
        type="submit"
        className="bg-blue-600 text-white px-4 py-2 rounded"
      >
        Crear empleado
      </button>
    </form>
  );
}
