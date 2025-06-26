declare module "test" {
    /**
     * Fails the test with the given message.
     * Equivalent to calling `t.Errorf` in Go.
     * @param msg The error message to display.
     */
    export function fail(msg: string): void;

    /**
     * Logs a message to the test output (non-fatal).
     * Equivalent to calling `t.Logf` in Go.
     * @param msg The message to log.
     */
    export function log(msg: string): void;
}
