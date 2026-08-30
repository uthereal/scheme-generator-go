package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/uthereal/scheme-generator-go/internal/generator"
)

// main is the entry point for the scheme generator CLI application.
func main() {
	var inputDir string
	var pkgName string
	var outDir string

	flag.StringVar(&inputDir, "dir", "", "Input folder of SQL files")
	flag.StringVar(&pkgName, "package", "", "Go package name for output")
	flag.StringVar(&outDir, "output", "", "Output directory")
	flag.Parse()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	if inputDir == "" || pkgName == "" || outDir == "" {
		logger.ErrorContext(
			ctx,
			"Invalid CLI arguments provided.",
			slog.String("dir", inputDir),
			slog.String("package", pkgName),
			slog.String("output", outDir),
		)
		os.Exit(1)
	}

	fSys := os.DirFS(inputDir)

	err := generator.Run(fSys, pkgName, outDir)
	if err != nil {
		logger.ErrorContext(
			ctx,
			"Failed to generate code.",
			slog.Any("error", err),
			slog.String("package", pkgName),
		)
		os.Exit(1)
	}

	logger.InfoContext(
		ctx,
		"Successfully generated code for package.",
		slog.String("package", pkgName),
	)
}
