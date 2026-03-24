import { redirect } from 'next/navigation';

export default function EpisodePage({ params }: { params: { pid: string; sid: string; eid: string } }) {
  redirect(`/projects/${params.pid}/seasons/${params.sid}/episodes/${params.eid}/script`);
}
