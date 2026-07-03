document.addEventListener("DOMContentLoaded", () => {
    /*
    ==========================
        Reveal Animation
    ==========================
    */
    const revealElements = document.querySelectorAll(
        ".feature-card, .product-card, .step, .review-card, .faq-item, .stat-item"
    );
    revealElements.forEach(el => {
        el.classList.add("reveal");
    });
    const observer = new IntersectionObserver(entries => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add("active");
            }
        });
    }, {
        threshold: 0.15
    });
    revealElements.forEach(el => observer.observe(el));

    /*
    ==========================
        Counter Animation
    ==========================
    */
    document.querySelectorAll("[data-counter]").forEach(counter => {
        const target = Number(counter.dataset.counter);
        let current = 0;
        const step = Math.ceil(target / 100);
        const timer = setInterval(() => {
            current += step;
            if (current >= target) {
                current = target;
                clearInterval(timer);
            }
            counter.textContent = current.toLocaleString("ru-RU") + "+";
        }, 15);
    });

    /*
    ==========================
        Hero Parallax
    ==========================
    */
    const hero = document.querySelector(".hero-wrapper");
    window.addEventListener("scroll", () => {
        const offset = window.pageYOffset;
        hero.style.backgroundPositionY = offset * 0.4 + "px";
    });

    /*
    ==========================
        Mouse Glow Effect
    ==========================
    */
    document.querySelectorAll(".feature-card,.product-card,.review-card,.step")
    .forEach(card => {
        card.addEventListener("mousemove", (e) => {
            const rect = card.getBoundingClientRect();
            const x = e.clientX - rect.left;
            const y = e.clientY - rect.top;
            card.style.background =
                `radial-gradient(circle at ${x}px ${y}px, rgba(255,255,255,.10), rgba(255,255,255,.04) 45%)`;
        });
        card.addEventListener("mouseleave", () => {
            card.style.background = "rgba(255,255,255,.04)";
        });
    });

    /*
    ==========================
        Sticky Header Shadow
    ==========================
    */
    const header = document.querySelector("header");
    window.addEventListener("scroll", () => {
        if (window.scrollY > 40) {
            header.style.boxShadow = "0 10px 35px rgba(0,0,0,.45)";
        } else {
            header.style.boxShadow = "none";
        }
    });

    /*
    ==========================
        Button Ripple
    ==========================
    */
    document.querySelectorAll(".btn").forEach(btn => {
        btn.addEventListener("click", function(e) {
            const circle = document.createElement("span");
            const diameter = Math.max(this.clientWidth, this.clientHeight);
            const radius = diameter / 2;
            circle.style.width = circle.style.height = diameter + "px";
            circle.style.left = e.clientX - this.getBoundingClientRect().left - radius + "px";
            circle.style.top = e.clientY - this.getBoundingClientRect().top - radius + "px";
            circle.classList.add("ripple");
            const ripple = this.querySelector(".ripple");
            if (ripple) {
                ripple.remove();
            }
            this.appendChild(circle);
        });
    });
});