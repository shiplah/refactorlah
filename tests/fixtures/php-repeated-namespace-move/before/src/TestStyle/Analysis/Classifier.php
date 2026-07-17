<?php
namespace App\TestStyle\Analysis;

use External\Syntax\Expr;
use External\Syntax\Stmt;

enum ClassificationSafety
{
    case AlreadyCanonical;
    case Canonicalizable;
}

enum OutputPartKind
{
    case ExceptionMessage;
}

final readonly class Classifier
{
    public function classify(Stmt\Catch_ $catch): ClassificationSafety
    {
        if ($catch->var instanceof Expr\Variable) {
            return ClassificationSafety::AlreadyCanonical;
        }

        OutputPartKind::ExceptionMessage;

        return ClassificationSafety::Canonicalizable;
    }
}
