import { PortfolioDetailSlice } from "@/features/portfolio/components/PortfolioDetailSlice";

type PortfolioPageProps = {
  params: Promise<{ portfolioId: string }>;
};

export default async function PortfolioPage({ params }: PortfolioPageProps) {
  const { portfolioId } = await params;
  return <PortfolioDetailSlice portfolioId={portfolioId} />;
}
