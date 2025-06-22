// frontend/components/employee-delete.jsx
"use client";

import React, { useEffect, useState } from "react";
import { toast } from "sonner";

export default function DeleteEmployeeForm({ onDeleted }) {
  const [employees, setEmployees] = useState([]);
  const [employeeId, setEmployeeId] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Carga lista de empleados
    fetch("http://localhost:3000/employees", {
      credentials: "include",
    })
      .then((res) => {
        if (!res.ok) throw new Error("Error cargando empleados");
        return res.json();
      })
      .then((data) => setEmployees(data))
      .catch((err) => {
        console.error(err);
        toast.error("No se pudieron cargar los empleados");
      })
      .finally(() => setLoading(false));
  }, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!employeeId) {
      toast.error("Selecciona un empleado válido");
      return;
    }
    if (!confirm("¿Seguro que quieres eliminar este empleado?")) return;

    try {
      const res = await fetch(`http://localhost:3000/employees/${employeeId}`, {
        method: "DELETE",
        credentials: "include",
      });
      if (res.status === 204) {
        toast.success("Empleado eliminado correctamente");
        // Actualiza la lista local
        setEmployees((prev) => prev.filter((emp) => emp.id !== parseInt(employeeId, 10)));
        setEmployeeId("");
        onDeleted?.(employeeId);
      } else {
        const errText = await res.text();
        console.error("Error servidor:", errText);
        throw new Error("Error al eliminar empleado");
      }
    } catch (err) {
      console.error(err);
      toast.error(err.message);
    }
  };

  if (loading) return <p>Cargando empleados…</p>;

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4 w-full mt-8 max-w-md">
      <label className="font-semibold">Selecciona un empleado a eliminar:</label>
      <select
        value={employeeId}
        onChange={(e) => setEmployeeId(e.target.value)}
        required
        className="border rounded p-2"
      >
        <option value="">-- Selecciona empleado --</option>
        {employees.map((emp) => (
          <option key={emp.id} value={emp.id}>
            {emp.name} (ID: {emp.id})
          </option>
        ))}
      </select>

      <button
        type="submit"
        className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded"
      >
        Eliminar empleado
      </button>
    </form>
  );
}
