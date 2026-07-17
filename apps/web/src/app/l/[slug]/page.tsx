import { Board } from '@/components/board';

export default async function ListPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <Board initialSlug={slug} />;
}
