import { Navigate } from "react-router-dom";
import { useAuth } from "@clerk/react";
import { Navbar } from "../components/landing/Navbar";
import { Hero } from "../components/landing/Hero";
import { StoryBanner } from "../components/landing/StoryBanner";
import { HowItWorks } from "../components/landing/HowItWorks";
import { Showcase } from "../components/landing/Showcase";
import { Features } from "../components/landing/Features";
import { Telemetry } from "../components/landing/Telemetry";
import { Audience } from "../components/landing/Audience";
import { EmotionalBanner } from "../components/landing/EmotionalBanner";
import { FinalCTA } from "../components/landing/FinalCTA";
import { Footer } from "../components/landing/Footer";
import { ScrollToTop } from "../components/ui/ScrollToTop";

export function Landing() {
  const { isLoaded, isSignedIn } = useAuth();

  if (isLoaded && isSignedIn) {
    return <Navigate to="/enter" replace />;
  }

  return (
    <div className="min-h-dvh overflow-x-clip bg-[#131315] font-body-md text-on-surface antialiased">
      <Navbar />
      <main className="w-full bg-[#131315] pt-16">
        <Hero />
        <StoryBanner />
        <HowItWorks />
        <Showcase />
        <Features />
        <Telemetry />
        <Audience />
        <EmotionalBanner />
        <FinalCTA />
      </main>
      <Footer />
      <ScrollToTop />
    </div>
  );
}
