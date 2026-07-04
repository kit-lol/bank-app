document.addEventListener("DOMContentLoaded", () => {
    /*
    ==========================
        Toast Notifications (Уведомления)
    ==========================
    */
    window.showToast = function(message, type = 'error') {
        let container = document.getElementById('toast-container');
        
        // Если контейнера нет в HTML, создаем его динамически
        if (!container) {
            container = document.createElement('div');
            container.id = 'toast-container';
            container.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 9999; display: flex; flex-direction: column; gap: 10px;';
            document.body.appendChild(container);
        }

        const toast = document.createElement('div');
        toast.className = `toast ${type}`;
        
        // Иконки
        let icon = '⚠️';
        if (type === 'success') icon = '✅';
        if (type === 'error') icon = '❌';

        toast.innerHTML = `
            <div style="display:flex; align-items:center; gap:10px;">
                <span style="font-size:18px;">${icon}</span>
                <span>${message}</span>
            </div>
            <button class="toast-close" style="background:none; border:none; color:#aaa; cursor:pointer; font-size:18px; margin-left:10px;">×</button>
        `;

        // Логика закрытия
        toast.querySelector('.toast-close').onclick = () => removeToast(toast);
        
        container.appendChild(toast);

        // Авто-удаление через 5 секунд
        setTimeout(() => removeToast(toast), 5000);
    };

    function removeToast(toast) {
        toast.style.opacity = '0';
        toast.style.transform = 'translateX(100%)';
        setTimeout(() => toast.remove(), 300);
    }

    /*
    ==========================
        Reveal Animation (Появление элементов)
    ==========================
    */
    const revealElements = document.querySelectorAll(
        ".feature-card, .product-card, .step, .review-card, .faq-item, .stat-item"
    );
    
    // Добавляем класс для начального скрытия
    revealElements.forEach(el => el.classList.add("reveal"));

    const observer = new IntersectionObserver(entries => {
        entries.forEach(entry => {
            if (entry.isIntersecting) {
                entry.target.classList.add("active");
            }
        });
    }, { threshold: 0.15 });

    revealElements.forEach(el => observer.observe(el));

    /*
    ==========================
        Counter Animation (Счетчики)
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
        Hero Parallax (Параллакс фона)
    ==========================
    */
    const hero = document.querySelector(".hero-wrapper");
    if (hero) {
        window.addEventListener("scroll", () => {
            const offset = window.pageYOffset;
            hero.style.backgroundPositionY = offset * 0.4 + "px";
        });
    }

    /*
    ==========================
        Mouse Glow Effect (Свечение за мышкой)
    ==========================
    */
    document.querySelectorAll(".feature-card, .product-card, .review-card, .step")
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
        Sticky Header Shadow (Тень шапки)
    ==========================
    */
    const header = document.querySelector("header");
    if (header) {
        window.addEventListener("scroll", () => {
            if (window.scrollY > 40) {
                header.style.boxShadow = "0 10px 35px rgba(0,0,0,.45)";
            } else {
                header.style.boxShadow = "none";
            }
        });
    }

    /*
    ==========================
        Button Ripple (Рябь на кнопках)
    ==========================
    */
    document.querySelectorAll(".btn").forEach(btn => {
        btn.addEventListener("click", function(e) {
            const circle = document.createElement("span");
            const diameter = Math.max(this.clientWidth, this.clientHeight);
            const radius = diameter / 2;
            
            circle.style.width = circle.style.height = `${diameter}px`;
            circle.style.left = `${e.clientX - this.getBoundingClientRect().left - radius}px`;
            circle.style.top = `${e.clientY - this.getBoundingClientRect().top - radius}px`;
            circle.classList.add("ripple");
            
            const existingRipple = this.querySelector(".ripple");
            if (existingRipple) existingRipple.remove();
            
            this.appendChild(circle);
        });
    });
});