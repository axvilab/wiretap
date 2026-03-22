<?php

if (! function_exists('wiretap')) {
    /**
     * Silently send a value to Wiretap for inspection.
     * Unlike dump(), this produces no output — the value is only visible in Wiretap UI.
     *
     * @param  mixed       $value  The value to inspect
     * @param  string|null $label  Optional label to identify this dump in the trace tree
     */
    function wiretap(mixed $value, ?string $label = null): mixed
    {
        if (\Axvi\Wiretap\Wiretap::$rid !== '') {
            $trace = debug_backtrace(DEBUG_BACKTRACE_IGNORE_ARGS, 2);
            $caller = $trace[0];

            \Axvi\Wiretap\Wiretap::send('wiretap', [
                'value' => print_r($value, true),
                'label' => $label,
                'file' => $caller['file'] ?? '',
                'line' => $caller['line'] ?? 0,
            ]);
        }

        return $value;
    }
}
