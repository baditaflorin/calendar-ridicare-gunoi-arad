package etl

import (
	"testing"
)

func TestCleanCartier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Micalaca variations
		{"Micalaca", "Micalaca"},
		{"Micala Stanga", "Micalaca"},
		{"Micalaca Dreapta", "Micalaca"},
		{"Micalaca dreapta", "Micalaca"},
		{"Micălaca – Zona 100-200", "Micalaca"},
		{"Micălaca – Zona 300", "Micalaca"},
		{"Micălaca – Zona 500-700", "Micalaca"},
		
		// Gradiste variations
		{"Gradiste", "Gradiste"},
		{"Gradiste Dreapta", "Gradiste"},
		{"Gradiste Stanga", "Gradiste"},
		{"Gradiste dr", "Gradiste"},
		{"Gradiste stg", "Gradiste"},
		{"Grădiște – partea dreaptă", "Gradiste"},
		{"Grădiște – partea stângă", "Gradiste"},

		// Aradul Nou variations
		{"Aradul Nou Dr", "Aradul Nou"},
		{"Aradul Nou Stg", "Aradul Nou"},
		{"Aradul Nou – partea dreaptă", "Aradul Nou"},
		{"Aradul Nou – partea stângă", "Aradul Nou"},
		{"Aradul nou", "Aradul Nou"},

		// Centru variations
		{"Centru", "Centru"},
		{"Centru – partea dreaptă", "Centru"},
		{"Centru – partea stângă", "Centru"},

		// Vanatori variations
		{"6 Vanatori", "6 Vanatori"},
		{"6 Vânători", "6 Vanatori"},
		{"6 vanatori", "6 Vanatori"},

		// Parneava variations
		{"Parneava", "Parneava"},
		{"Pârneava", "Parneava"},

		// Dragasani variations
		{"Dragasani", "Dragasani"},
		{"Drăgășani", "Dragasani"},

		// Sanicolau Mic variations
		{"Sanicolau Mic", "Sanicolau Mic"},
		{"Sanicolau mic", "Sanicolau Mic"},
		{"Sânnicolaul Mic", "Sanicolau Mic"},

		// Romana Residence variations
		{"Romana Residence", "Romana Residence"},
		{"Romana Residence – privat", "Romana Residence"},

		// Verde variations
		{"Verde – privat", "Verde"},
		{"verde", "Verde"},
		{"Verde", "Verde"},

		// Vlaicu variations
		{"Vlaicu", "Aurel Vlaicu"},
		{"Vlaicu dreapta", "Aurel Vlaicu"},

		// Others without variations that should remain unchanged
		{"Alfa", "Alfa"},
		{"Bujac", "Bujac"},
		{"Cadaș Silvaș", "Cadaș Silvaș"},
		{"Confectii", "Confectii"},
		{"Gai", "Gai"},
		{"I. G. Duca", "I. G. Duca"},
		{"Insula Mureș", "Insula Mureș"},
		{"Mandarinilor", "Mandarinilor"},
		{"Muresel", "Muresel"},
		{"Poltura", "Poltura"},
		{"San Paolo", "San Paolo"},
		{"Sega", "Sega"},
		{"Subcetate", "Subcetate"},
		{"Trupuri izolate de intravilan", "Trupuri izolate de intravilan"},
		{"Westfield", "Westfield"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			actual := cleanCartier(tt.input)
			if actual != tt.expected {
				t.Errorf("cleanCartier(%q) = %q; expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}
