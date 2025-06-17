import { useMemo } from "react";

export function useOrdinal(day: number): string {
    return useMemo(() => {
        if (day >= 11 && day <= 13) {
            return `${day}th`;
        }
        
        const lastDigit = day % 10;
        switch (lastDigit) {
            case 1:
                return `${day}st`;
            case 2:
                return `${day}nd`;
            case 3:
                return `${day}rd`;
            default:
                return `${day}th`;
        }
    }, [day]);
}

// Also export as a standalone utility function
export function toOrdinal(day: number): string {
    if (day >= 11 && day <= 13) {
        return `${day}th`;
    }
    
    const lastDigit = day % 10;
    switch (lastDigit) {
        case 1:
            return `${day}st`;
        case 2:
            return `${day}nd`;
        case 3:
            return `${day}rd`;
        default:
            return `${day}th`;
    }
}