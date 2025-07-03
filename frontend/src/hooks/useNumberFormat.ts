export function useNumberFormat() {
    function formatFloat(input: number): number {
        const fixed = input.toFixed(2);
        return parseFloat(fixed);
    }

    return {
        formatFloat,
    };
}
