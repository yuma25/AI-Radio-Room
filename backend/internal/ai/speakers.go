/**
 * Package ai 内の話者管理ロジックを定義する。
 */
package ai

/**
 * VOICEVOX の話者ID（スタイルID）を定数として定義する。
 */
const (
	SpeakerMetanNormal  = 2  // 四国めたん（ノーマル）
	SpeakerMetanExcited = 6  // 四国めたん（ツンツン）
	SpeakerMetanCalm    = 36 // 四国めたん（ささやき）

	SpeakerZundaNormal  = 3  // ずんだもん（ノーマル）
	SpeakerZundaExcited = 7  // ずんだもん（ツンツン）
	SpeakerZundaCalm    = 22 // ずんだもん（ささやき）

	SpeakerTsumugiNormal = 8 // 春日部つむぎ（ノーマル）
)

/**
 * GetSpeakerID は、役割名と感情から最適な VOICEVOX スタイルIDを返す。
 * @param role [string] mc, guest-a, guest-b
 * @param emotion [string] normal, excited, calm
 * @return [int] VOICEVOX スタイルID
 */
func GetSpeakerID(role, emotion string) int {
	switch role {
	case "mc":
		switch emotion {
		case "excited":
			return SpeakerMetanExcited
		case "calm":
			return SpeakerMetanCalm
		default:
			return SpeakerMetanNormal
		}
	case "guest-a":
		switch emotion {
		case "excited":
			return SpeakerZundaExcited
		case "calm":
			return SpeakerZundaCalm
		default:
			return SpeakerZundaNormal
		}
	case "guest-b":
		return SpeakerTsumugiNormal
	default:
		return SpeakerMetanNormal
	}
}
