<?php
final class NodeRenderer
{
    public function __construct(private iterable $componentRenderers) {}

    public function tag(): string
    {
        return 'app.component_renderer';
    }
}
