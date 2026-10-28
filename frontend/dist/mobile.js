// Keep chat and sheets inside the visible viewport when a mobile keyboard opens.
export function setupMobileViewport() {
    const root = document.documentElement;
    const viewport = window.visualViewport;
    let width = window.innerWidth;
    let referenceHeight = window.innerHeight;
    let frame;
    const update = () => {
        const height = viewport?.height || window.innerHeight;
        if (Math.abs(window.innerWidth - width) > 50) {
            width = window.innerWidth;
            referenceHeight = window.innerHeight;
        }
        referenceHeight = Math.max(referenceHeight, window.innerHeight, height);
        const editing = document.activeElement?.matches('textarea, input:not([type="checkbox"]):not([type="radio"]), select');
        const keyboard = window.matchMedia('(max-width: 760px), (max-width:1000px) and (max-height:540px) and (pointer:coarse)').matches && editing && height < referenceHeight * .78;
        root.style.setProperty('--app-height', `${Math.round(height)}px`);
        root.style.setProperty('--viewport-top', `${Math.round(viewport?.offsetTop || 0)}px`);
        root.classList.toggle('keyboard-open', Boolean(keyboard));
    };
    const schedule = () => { cancelAnimationFrame(frame); frame = requestAnimationFrame(update); };
    window.addEventListener('resize', schedule);
    viewport?.addEventListener('resize', schedule);
    viewport?.addEventListener('scroll', schedule);
    document.addEventListener('focusin', schedule);
    document.addEventListener('focusout', schedule);
    update();
}
