package records

import "testing"

func TestImportValidateAndDeduplicate(t *testing.T) {
	service := NewService()
	request := ImportRequest{Name: "示例", Format: "iccs", Content: "a3a4 a6a5"}
	first := service.Import(request)
	if first.ImportedGames != 1 || first.FailedGames != 0 {
		t.Fatalf("first import: %+v", first)
	}
	record, err := service.Get(first.RecordIDs[0])
	if err != nil {
		t.Fatal(err)
	}
	if record.MoveCount != 2 || record.Moves[1].Side != "black" {
		t.Fatalf("record: %+v", record)
	}
	second := service.Import(request)
	if second.DuplicateGames != 1 || second.RecordIDs[0] != record.ID {
		t.Fatalf("duplicate import: %+v", second)
	}
}

func TestImportRejectsIllegalMoveWithPly(t *testing.T) {
	batch := NewService().Import(ImportRequest{Format: "iccs", Content: "a3a2"})
	if batch.FailedGames != 1 || len(batch.Errors) != 1 {
		t.Fatalf("batch: %+v", batch)
	}
	if batch.Errors[0].Ply != 1 || batch.Errors[0].Code != "ILLEGAL_MOVE" {
		t.Fatalf("error: %+v", batch.Errors[0])
	}
}
