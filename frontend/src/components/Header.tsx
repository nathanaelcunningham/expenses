import { Link } from "@tanstack/react-router";

export default function Header() {
    return (
        <nav className="border-b py-3">
            <div className="mx-auto px-4 flex justify-between items-center">
                <Link className="text-lg font-bold" to="/">
                    Home
                </Link>
            </div>
        </nav>
    );
}
