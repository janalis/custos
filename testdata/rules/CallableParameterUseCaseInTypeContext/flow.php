<?php
class Writer
{
    private function next(string $doc): ?string { return null; }

    public function insert(string $doc, array $items): string
    {
        foreach ($items as $item) {
            $next = $this->next($doc);
            // narrowed by the guard: a string
            if (null !== $next) {
                $doc = $next;
            }
        }
        return $doc;
    }

    public function hosts(array|string|null $hosts): void
    {
        if (is_string($hosts)) {
            // `?:` only yields its left operand when truthy (never false)
            $hosts = preg_split('/\R/', $hosts) ?: [];
        }
    }

    public function reset(string $doc): string
    {
        $next = $this->next($doc);
        $doc = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$next</warning>;
        return (string) $doc;
    }
}
