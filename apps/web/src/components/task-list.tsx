'use client';

import {
  DndContext,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core';
import {
  SortableContext,
  arrayMove,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import type { Task } from '@/lib/api/lists';
import { useDeleteTask, useEditTask, useReorderTasks, useToggleTask } from '@/hooks/use-list';
import { TaskItem } from '@/components/task-item';

interface TaskListProps {
  slug: string;
  color: string;
  tasks: Task[];
}

export function TaskList({ slug, color, tasks }: TaskListProps) {
  const toggleTask = useToggleTask(slug);
  const editTask = useEditTask(slug);
  const deleteTask = useDeleteTask(slug);
  const reorderTasks = useReorderTasks(slug);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  );

  const handleDragEnd = (event: DragEndEvent) => {
    const { active, over } = event;
    if (!over || active.id === over.id) return;
    const oldIndex = tasks.findIndex((t) => t.id === active.id);
    const newIndex = tasks.findIndex((t) => t.id === over.id);
    if (oldIndex < 0 || newIndex < 0) return;
    const newOrder = arrayMove(tasks, oldIndex, newIndex).map((t) => t.id);
    reorderTasks.mutate(newOrder);
  };

  if (tasks.length === 0) {
    return <p className="text-sm text-muted-foreground">Nenhuma tarefa ainda.</p>;
  }

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
      <SortableContext items={tasks.map((t) => t.id)} strategy={verticalListSortingStrategy}>
        <ul className="space-y-2">
          {tasks.map((task) => (
            <TaskItem
              key={task.id}
              task={task}
              color={color}
              onToggle={(done) => toggleTask.mutate({ taskId: task.id, done })}
              onEdit={(title) => editTask.mutate({ taskId: task.id, title })}
              onDelete={() => deleteTask.mutate(task.id)}
            />
          ))}
        </ul>
      </SortableContext>
    </DndContext>
  );
}
