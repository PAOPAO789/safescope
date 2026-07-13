import { ProjectDetailView } from "@/components/project-detail-view";

export default async function ProjectPage({ params }: { params: Promise<{ id: string }> }) {
  return <ProjectDetailView id={(await params).id} />;
}
