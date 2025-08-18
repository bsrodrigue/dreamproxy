document.addEventListener("DOMContentLoaded", () => {
  // Navigation toggle
  const navToggle = document.getElementById("navToggle");
  const nav = document.getElementById("primaryNav");
  const header = document.querySelector(".site-header");

  navToggle.addEventListener("click", () => {
    const isOpen = nav.getAttribute("data-open") === "true";
    nav.setAttribute("data-open", !isOpen);
    navToggle.setAttribute("aria-expanded", !isOpen);
  });

  // Smooth scroll for anchor links
  document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener("click", (e) => {
      e.preventDefault();
      const target = document.querySelector(anchor.getAttribute("href"));
      if (target) {
        target.scrollIntoView({ behavior: "smooth" });
      }
    });
  });

  // Scroll-based header styling
  window.addEventListener("scroll", () => {
    header.setAttribute("data-scrolled", window.scrollY > 50);
  });

  // Reveal animations
  const reveals = document.querySelectorAll(".reveal");
  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          entry.target.classList.add("reveal");
          observer.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.2 }
  );

  reveals.forEach((el) => observer.observe(el));

  // Form handling
  const form = document.querySelector(".signup");
  const formMsg = document.querySelector(".form-msg");

  form.addEventListener("submit", (e) => {
    e.preventDefault();
    const email = form.querySelector("#email").value;
    formMsg.textContent = `Thanks for signing up with ${email}! We'll reach out soon.`;
    formMsg.style.color = "#fff";
    form.reset();
  });

  // Set current year in footer
  document.getElementById("year").textContent = new Date().getFullYear();
});
