import React, { useEffect, useRef } from "react";
import { gsap } from "gsap";
import { Link } from "react-router";
import styles from "./Home.module.css";

const Home = () => {
  const heroRef = useRef<HTMLDivElement>(null);
  const titleRef = useRef<HTMLHeadingElement>(null);
  const subtitleRef = useRef<HTMLParagraphElement>(null);
  const btnRef = useRef<HTMLAnchorElement>(null);

  useEffect(() => {
    const tl = gsap.timeline({ defaults: { ease: "power2.out", duration: 1 } });
    tl.from(heroRef.current, { opacity: 0, duration: 1 })
      .from(titleRef.current, { y: -50, opacity: 0 }, "-=0.5")
      .from(subtitleRef.current, { y: 50, opacity: 0 }, "-=0.75")
      .from(btnRef.current, { scale: 0.8, opacity: 0 }, "-=0.5");
  }, []);

  return (
    <div className={styles.home} ref={heroRef}>
      <h1 ref={titleRef} className={styles.title}>
        Добро пожаловать в Lumivy!
      </h1>
      <p ref={subtitleRef} className={styles.subtitle}>
        Погрузитесь в мир удобного управления событиями и встречами.
      </p>
      <Link ref={btnRef} to="/dashboard" className={styles.ctaButton}>
        Перейти к платформе
      </Link>
    </div>
  );
};

export default Home;
