import { ref } from 'vue'

const chineseChars = '的一是在不了有和人这中大为上个国我以要他时来用们生到作地于出就分对成会可主发年动同工也能下过子说产种面而方后多定行学法所民得经十三之进着等部度家电力里如水化高自二理起小物现实加量都两体制机当使点从业本去把性好应开它合还因由其些然前外天政四日那社义事平形相全表间样与关各重新线内数正心反你明看原又么利比或但质气第向道命此变条只没结解问意建月公无系军很情者最立代想已通并提直题党程展五果料象员革位入常文总次品式活设及管特件长求老头基资边流路级少图山统接知较将组见计别她手角期根论运农指几九区强放决西被干做必战先回则任取据处队南给色光门即保治北造百规热领七海口东导器压志世金增争济阶油思术极交受联什认六共权收证改清己美再采转更单风切打白教速花带安场身车例真务具万每目至达走积示议声报斗完类八离华名确才科张信马节话米整空元况今集温传土许步群广石记需段研界拉林律叫且究观越织装影算低持音众书布复容儿须际商非验连断深难近矿千周委素技备半办青省列习响约支般史感劳便团往酸历市克何除消构府称太准精值号率族维划选标写存候毛亲快效斯院查江型眼王按格养易置派层片始却专状育厂京识适属圆包火住调满县局照参红细引听该铁价严龙飞'

const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'

function shuffle<T>(arr: T[]): T[] {
  const a = [...arr]
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[a[i], a[j]] = [a[j], a[i]]
  }
  return a
}

function rand(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

export type CaptchaType = 'math' | 'category' | 'shape' | 'chinese' | 'letter'

export interface CaptchaChallenge {
  type: CaptchaType
  question: string
  answer: string
  /** For shape/click types: items to display for selection */
  options?: { id: string; label: string }[]
}

function generateChallenge(): CaptchaChallenge {
  const typeNum = rand(1, 5)

  switch (typeNum) {
    // 1. 数学 - 四则运算
    case 1: {
      const op = rand(1, 4)
      switch (op) {
        case 1: { const a = rand(1,20), b = rand(1,20); return { type:'math', question:`${a} + ${b} = ?`, answer:String(a+b) } }
        case 2: { const a = rand(5,30), b = rand(1,a); return { type:'math', question:`${a} - ${b} = ?`, answer:String(a-b) } }
        case 3: { const a = rand(2,9), b = rand(2,9); return { type:'math', question:`${a} × ${b} = ?`, answer:String(a*b) } }
        case 4: { const a = rand(2,20), b = rand(2,9), r = Math.floor(a/b), rm = a % b; return { type:'math', question:`${a} ÷ ${b} = ?（取整）`, answer:String(r) } }
      }
    }

    // 2. 分类选择 - 从选项中选出指定类别
    case 2: {
      const categories: { q: string; opts: { label: string; correct: boolean }[] }[] = [
        { q: '下列哪个是颜色？', opts: [
          { label: '跑步', correct: false }, { label: '蓝色', correct: true },
          { label: '吃饭', correct: false }, { label: '睡觉', correct: false },
        ]},
        { q: '下列哪个是数字？', opts: [
          { label: '苹果', correct: false }, { label: '七', correct: true },
          { label: '太阳', correct: false }, { label: '河流', correct: false },
        ]},
        { q: '下列哪个是动物？', opts: [
          { label: '桌子', correct: false }, { label: '椅子', correct: false },
          { label: '猫咪', correct: true }, { label: '窗户', correct: false },
        ]},
        { q: '下列哪个是水果？', opts: [
          { label: '香蕉', correct: true }, { label: '铅笔', correct: false },
          { label: '书包', correct: false }, { label: '电视', correct: false },
        ]},
        { q: '下列哪个是交通工具？', opts: [
          { label: '冰箱', correct: false }, { label: '沙发', correct: false },
          { label: '飞机', correct: true }, { label: '牙刷', correct: false },
        ]},
        { q: '下列哪个不属于电器？', opts: [
          { label: '冰箱', correct: false }, { label: '洗衣机', correct: false },
          { label: '石头', correct: true }, { label: '空调', correct: false },
        ]},
        { q: '下列哪个是乐器？', opts: [
          { label: '钢笔', correct: false }, { label: '吉他', correct: true },
          { label: '毛巾', correct: false }, { label: '水杯', correct: false },
        ]},
      ]
      const pick = categories[rand(0, categories.length - 1)]
      const correctOpt = pick.opts.find(o => o.correct)!
      const options = shuffle(pick.opts.map(o => ({ id: o.label, label: o.label })))
      return { type:'category', question: pick.q, answer: correctOpt.label, options }
    }

    // 3. 形状 - 从多个图形中选出指定的
    case 3: {
      const shapes = [
        { id:'○', label:'圆形' }, { id:'□', label:'正方形' },
        { id:'△', label:'三角形' }, { id:'☆', label:'星形' },
        { id:'◇', label:'菱形' }, { id:'❤', label:'心形' },
      ]
      const pick = shapes[rand(0, shapes.length - 1)]
      const distractors = shuffle(shapes.filter(s => s.id !== pick.id)).slice(0, 3)
      const options = shuffle([pick, ...distractors])
      return { type:'shape', question:`请点击 "${pick.label}"`, answer: pick.id, options: options.map(s => ({ id: s.id, label: s.id })) }
    }

    // 4. 顺序点击中文字符
    case 4: {
      const count = rand(2, 4)
      const targets: string[] = []
      while (targets.length < count) {
        const c = chineseChars[rand(0, chineseChars.length - 1)]
        if (!targets.includes(c)) targets.push(c)
      }
      const answer = targets.join(',')
      const optSet = new Set(targets)
      while (optSet.size < 8) {
        optSet.add(chineseChars[rand(0, chineseChars.length - 1)])
      }
      const options = shuffle(Array.from(optSet).map(c => ({ id: c, label: c })))
      return { type:'chinese', question:`请按顺序点击：${targets.join(' → ')}`, answer, options }
    }

    // 5. 顺序点击英文字符
    case 5: {
      const count = rand(2, 4)
      const targets: string[] = []
      const usedIdx = new Set<number>()
      while (targets.length < count) {
        const idx = rand(0, 25)
        if (!usedIdx.has(idx)) { usedIdx.add(idx); targets.push(letters[idx]) }
      }
      targets.sort((a, b) => a.localeCompare(b))
      const answer = targets.join(',')
      // 构建选项：必须包含 targets 中的所有字母，再加一些干扰项
      const optSet = new Set(targets)
      while (optSet.size < 8) {
        optSet.add(letters[rand(0, 25)])
      }
      const options = shuffle(Array.from(optSet).map(l => ({ id: l, label: l })))
      return { type:'letter', question:`请按字母顺序点击：${targets.join(' → ')}`, answer, options }
    }
  }
}

export function useCaptcha() {
  const challenge = ref<CaptchaChallenge>({ type: 'math', question: '', answer: '' })
  const userAnswer = ref('')
  const captchaText = ref('')
  const clickedIds = ref<string[]>([])

  function refresh() {
    challenge.value = generateChallenge()
    captchaText.value = challenge.value.question
    userAnswer.value = ''
    clickedIds.value = []
  }

  function validate(): boolean {
    if (challenge.value.type === 'shape' || challenge.value.type === 'category') {
      return userAnswer.value === challenge.value.answer
    }
    if (challenge.value.type === 'chinese' || challenge.value.type === 'letter') {
      return clickedIds.value.join(',') === challenge.value.answer
    }
    // math
    return challenge.value.answer === userAnswer.value.trim()
  }

  function handleClick(id: string) {
    if (challenge.value.type === 'shape' || challenge.value.type === 'category') {
      userAnswer.value = id
      return
    }
    if (challenge.value.type === 'chinese' || challenge.value.type === 'letter') {
      if (clickedIds.value.includes(id)) return
      clickedIds.value.push(id)
      userAnswer.value = clickedIds.value.join(',')
    }
  }

  function isClicked(id: string): boolean {
    return clickedIds.value.includes(id)
  }

  refresh()

  return { challenge, userAnswer, captchaText, clickedIds, refresh, validate, handleClick, isClicked }
}
