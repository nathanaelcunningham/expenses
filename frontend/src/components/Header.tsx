import { Link } from "@tanstack/react-router";
import { useState } from "react";
import { useAuth } from "../contexts/AuthContext";

export default function Header() {
    const { user, isAuthenticated, logout } = useAuth();
    const [isUserMenuOpen, setIsUserMenuOpen] = useState(false);

    const handleLogout = async () => {
        await logout();
        setIsUserMenuOpen(false);
    };

    return (
        <nav className="border-b py-3 bg-white shadow-sm">
            <div className="mx-auto px-4 flex justify-between items-center">
                <Link className="text-lg font-bold text-blue-600" to="/">
                    Expense Tracker
                </Link>
                
                {isAuthenticated && user ? (
                    <div className="flex items-center space-x-4">
                        <Link 
                            to="/expense/create"
                            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium"
                        >
                            Add Expense
                        </Link>
                        
                        <div className="relative">
                            <button
                                onClick={() => setIsUserMenuOpen(!isUserMenuOpen)}
                                className="flex items-center space-x-2 text-gray-700 hover:text-gray-900 focus:outline-none"
                            >
                                <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
                                    <span className="text-sm font-medium text-blue-600">
                                        {user.name.charAt(0).toUpperCase()}
                                    </span>
                                </div>
                                <span className="text-sm font-medium">{user.name}</span>
                                <svg
                                    className={`w-4 h-4 transition-transform ${isUserMenuOpen ? 'rotate-180' : ''}`}
                                    fill="none"
                                    stroke="currentColor"
                                    viewBox="0 0 24 24"
                                >
                                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                                </svg>
                            </button>
                            
                            {isUserMenuOpen && (
                                <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 z-10 border">
                                    <div className="px-4 py-2 text-sm text-gray-700 border-b">
                                        <div className="font-medium">Signed in as</div>
                                        <div className="text-gray-600">{user.email}</div>
                                    </div>
                                    <button
                                        onClick={handleLogout}
                                        className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                                    >
                                        Sign out
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>
                ) : (
                    <div className="flex items-center space-x-4">
                        <Link
                            to="/login"
                            className="text-gray-600 hover:text-gray-900 text-sm font-medium"
                        >
                            Sign in
                        </Link>
                        <Link
                            to="/register"
                            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md text-sm font-medium"
                        >
                            Sign up
                        </Link>
                    </div>
                )}
            </div>
        </nav>
    );
}
