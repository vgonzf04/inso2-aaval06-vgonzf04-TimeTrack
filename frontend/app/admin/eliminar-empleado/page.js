"use client";

import EmployeeDeleteForm from "@/components/employee-delete";

export default function DeleteEmployeePage() {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Delete Employee</h1>
      <EmployeeDeleteForm />
    </div>
  );
}
